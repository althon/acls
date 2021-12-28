package acls

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Marshal(ptr interface{}) ([]byte,error){
	v:=reflect.ValueOf(ptr)
	return format_object(0,&v,true)
}

func Unmarshal(data []byte, ptr interface{}) error{
	value:=reflect.ValueOf(ptr)
	variable :=make(map[string]interface{})
	object,_:=parse_object(variable,0,0,0,data)

	o:=object.(map[string]interface{})
	for k,v:=range variable{
		o[k] = v
	}

	return set_object(&value,object)
}


func ToAcls(data []byte) Acls{
	variable := make(map[string]interface{})
	object,_:=parse_object(variable,0,0,0,data)
	o:=object.(map[string]interface{})
	for k,v:=range variable{
		o[k] = v
	}
	return o
}

func ToJson(data []byte) ([]byte,error){
	variable :=make(map[string]interface{})
	object,_:=parse_object(variable,0,0,0,data)
	if m,ok:= object.(map[string]interface{});ok{
		if m!=nil{
			for k,v:=range m{
				m[k] = v
			}
			return json.Marshal(m)
		}
	}
	return nil,errors.New("to json string fail")
}

func FromFile(filepath string) Acls{
	data,err:=os.ReadFile(filepath)
	if err!=nil{
		panic("cannot open file:" + filepath)
	}
	return ToAcls(bytes.TrimPrefix(data,[]byte{239, 187, 191}))
}
func parse_object(variable map[string]interface{}, main_level int,start int, end int,buf []byte) (interface{},int){
	var (
		filed_start = -1
		pos = start
		level = main_level
		size = end
		key = ""
		value  interface{}
		is_r = 0
		result interface{}
		is_eof = 0
	)

	if size == 0{
		size = len(buf)
	}

	for pos<size{
		switch buf[pos] {
		case '#':
			pos = skin_notes(pos+1,buf) //跳过
			level=0
		case ' ':
			level++
		case ':':
			if level!=main_level{
				panic(fmt.Sprintf("level is not different,near in '%s...'",buf[filed_start-10:pos+10]))
				//level = main_level
			}
			if result==nil{
				result = make(map[string]interface{})
			}

			key,value,pos = parse_element(variable,level,filed_start,pos,buf)
			result.(map[string]interface{})[key]=value
			filed_start,level = -1,0
			key,value="",nil
		case '\r':
			is_r = 1
		case '\n':
			if filed_start!=-1{
				panic(fmt.Sprintf("wrong field definition,near in '%s'",buf[pos:pos+10]))
			}
		case '-': //列表的下一个元素
		    if level+2>main_level{
				panic(fmt.Sprintf("level is not different,near in '%s...'",buf[pos:pos+10]))
			}else if level==0{
				panic(fmt.Sprintf("There must be at least one space before '-',near in '%s...'",buf[start:pos+10]))
			}
		    is_eof = 1
		default:
			if level<main_level{
				return result,pos-1 - level
			}
			if filed_start==-1{
				filed_start = pos
			}
		}
		if is_eof==1{
			break
		}
		pos++
	}

	if len(key)==0 && value==nil && result==nil{
		if filed_start==-1{
			result = string(buf[start:pos])
		}else{
			result = string(buf[filed_start:pos/*-1-is_r-level*/])
		}

	}
	if is_eof==1{ //It's list
		is_eof = 2
	}else{
		is_eof=0
	}
	return result,pos-level-is_r-is_eof
}

func parse_value(variable map[string]interface{},main_level int,start int,end int, buf []byte) (interface{},int){
	var(
		pos = start
		value_start_pos = -1 //'`'段落类型专用
		is_r = 0
		is_line = 0
		level = main_level
	)
	if end==0{
		end=len(buf)
	}
 	for pos<end{
		switch buf[pos] {
		case ' ':
			level++
		case '"','\'','`'://string type
			value_start_pos = pos+1-start
			value,pos:=parse_value_to_string(pos+1,end,value_start_pos,buf[pos],buf)
			return value,pos
		case '['://array type
			return parse_value_to_array(variable,main_level,pos + 1,0,buf)
		case '-'://list type
		    if level==0{
		    	panic(fmt.Sprintf("There must be at least one space before '-',near in '%s...'",buf[start:pos+10]))
			}
			return parse_value_to_list(variable,level,pos ,buf)
		case '\r':
			is_r=1
		case ':':
			//var value interface{}
			return parse_object(variable ,level,value_start_pos,0,buf)
		case '\n'://value type
			if value_start_pos!=-1 {
				return parse_value_to_valueType(value_start_pos,pos-is_r,buf),pos - is_r - 1
			}else{//key之后是换行符
				level=0
				is_line = 1
			}
		case '#':
			if value_start_pos==-1{
				pos = skin_notes(pos,buf)
				if level!=0{
					level=0
				}
			}
		default:
			if level == 0 && is_line==1{ //换行了，但层级是0
				panic(fmt.Sprintf("wrong format near in '%s...'",buf[pos:pos+10]))
			}

			if value_start_pos==-1 {
				value_start_pos = pos
			}
		}
		pos++
	}
	return  parse_value_to_valueType(start,end,buf),end-is_r
}
func parse_element(variable map[string]interface{},main_level int,start int, split_pos int,buf []byte) (string,interface{},int){
	var value interface{}
	alias:=""
	key,end := get_key(start,split_pos,buf),0
	alias,end = parse_variable(main_level,split_pos+1,buf)
	if end==0 || alias[0]!='$'{
		end = split_pos+1
		alias=""
	}else{
		alias="&" + alias[1:]
	}
	value,end = parse_value(variable,main_level,end,0,buf)
	if len(alias)>0{
		variable[alias]=value
	}else if var_name,ok:=value.(string);ok{
		if len(var_name)>0 && var_name[0]=='&'{
			if value,_ = variable[var_name];value==nil{
				panic("unavailable variable:" + var_name)
			}
		}
	}
	return key,value,end
}

func parse_variable(main_level,start int,buf []byte) (string,int){
	p:=  find_index_byte(start,buf,'$')
	if p!=-1{
		i,is_r,pos:=start,0,-1
		for ;i< len(buf);i++{
			if buf[i]=='$'{
				if i!=p{
					break
				}
			}
			switch buf[i] {
			case ' ':
				if pos!=-1{
					return string(buf[pos:i]),i
				}
			case '\r':
				is_r=1
			case '\n':
				if pos!=-1{
					return string(buf[pos:i-is_r]),i
				}
			default:
				if pos==-1{
					pos=i
				}
			}
		}
	}
    return "",0
}

func parse_value_to_string(start int,end int,para int, quot byte,buf []byte) (string,int){
	i,is_r,is_eof,size,text:=start,0,0,end,strings.Builder{}
	if size==0{
		size=len(buf)
	}
	for ;i< size;i++{
		switch buf[i] {
		case '\r':
			is_r=1
		case '\n'://遇到换行
			if buf[i-1-is_r]=='/'{ //如果是连接符
				text.Write(buf[start:i-1-is_r])
				start = skin_white(i+1,buf) //跳过'/'后面所有的空白处
				i = start-1
			}else if quot=='`'{ //段落
				if is_eof!=0{
					text.Write(buf[start : is_eof/*i-1-is_r*/])
					return text.String(),i
				}else{
					if buf[start]=='\r' || start==i{//此行只有换行符
						if is_r==1{
							text.WriteString("\r")
						}
						start = i+1
						para = 0
					}else if start!=i{
						text.Write(buf[start:i])
						start = i + 1
					}else{
						start = i+1+para //跳到下一个段落开始处
						i = start-1
					}
					text.WriteByte('\n')
				}
			}else if is_eof!=0{
				text.Write(buf[start : is_eof/*i-1-is_r*/])
				return text.String(),i - 1 - is_r
			}
			is_r=0
		case '\\':
			if buf[i+1]=='\\'{//如果下一个还是'\'
				i++ //直接跳过
			}
		case '"','\'','`'://出现疑似转义字符
			if buf[i]==quot {//如果结尾相同则表示该字符串结束或是转义字符
				if buf[i-1]=='\\'{ //如果是转义字符则转换
					text.Write(buf[start:i-1]) //截取
					text.WriteByte(quot)
					i++
					start=i
				}else{//字符串已经结束
					is_eof = i
				}
			}
		}
	}
	if is_eof!=0{ //没有换行符时，检测是否字符串结束
		text.Write(buf[start:is_eof])
		return text.String(),i
	}
	return "",-1
}

func parse_value_to_valueType(start int,end int, buf []byte) interface{}{
	start = skin_white(start,buf) //跳过空白
	end = end_white_bytes(0,buf[start:end])
	str:=string(buf[start:start+end])

	if strings.Index(str,".")!=-1{
		if value,err:= strconv.ParseFloat(str,64);err==nil{
			return value
		}else {
			return str
		}
	}else if value,err:= strconv.ParseInt(str,10,64);err==nil{
		return value
	}else if value,err:= strconv.ParseUint(str,10,64);err==nil{
		return value
	}else {
		switch strings.ToLower(str) {
		case "false":
			return false
		case "true":
			return true
		case "null":
			return nil
		default:
			return str
		}
	}
}

func parse_value_to_array(variable map[string]interface{},main_level int, start int,end int, buf []byte) (interface{},int){
	start = skin_white(start,buf)
	var (
		i = start
		//is_r = 0
		found = 0
		slice =make([]interface{},0)
		is_eof = 0
	)
	if end==0{
		end = len(buf)
	}
	found = bytes.LastIndexByte(buf[start:end],']')
	if found ==-1{
		panic(fmt.Sprintf("cannot find ']',near in the '%s...'",string(buf[start:end])))
	}

	end = start + found
	for ;i<=end;i++{
		switch buf[i] {
		case ',':
			if value,pos:=parse_value(variable,main_level,start,i,buf);pos!=-1{
				if var_name,ok:=value.(string);ok{
					if len(var_name)>0{
						if var_name[0]=='&'{
							if value,_ = variable[var_name];value==nil{
								panic("unavailable variable:" + var_name)
							}else{
								slice= append(slice, value.([]interface{})...)
							}
						}else{
							slice= append(slice, value)
						}
					}
				}else{
					slice= append(slice, value)
				}
				i=skin_white(i+1,buf)
				if buf[i]==']'{
					panic(fmt.Sprintf("range index error,near in '%s'",buf[start:i+1]))
				}
				start=i
			}
		case '\r','\n':
		//	is_r = 1
		//case '\n':
			//if value,pos:=get_value(start,i-is_r,buf);pos!=-1{
			//	slice= append(slice, value)
			//	i=skin_white(i+1,buf)
			//}
			//start=i-is_r
			//is_r = 0
		case ']'://结束
		    if start!=i{
				if value,pos:=parse_value(variable,main_level,start,i,buf);pos!=-1{
					if var_name,ok:=value.(string);ok{
						if len(var_name)>0{
							if var_name[0]=='&'{
								if value,_ = variable[var_name];value==nil{
									panic("unavailable variable:" + var_name)
								}else{
									slice= append(slice, value.([]interface{})...)
								}
							}else{
								slice= append(slice, value)
							}
						}

					}else{
						slice= append(slice, value)
					}
					//i=skin_white(i+1,buf)-1
				}
			}
			is_eof = 1
			//end= i + is_r + 1
		}

		if is_eof ==1{
			break
		}
	}
	//if buf[i-1]==']'{
	//	end= i
	//}
	return slice,i
}

func parse_value_to_list(variable map[string]interface{},main_level int,start int, buf []byte) (interface{},int){
	var value interface{}
	i,size,end,is_ele,level:=start,len(buf),0,0,main_level
	var result []interface{}
	for ;i<size;i++{
		switch buf[i] {
		case '-'://行
			if buf[i+1]!=' '{
				panic(fmt.Sprintf("it's not a element,near in '%s'...",string(buf[i:i+10])))
			}
			if level < main_level{ //已经不在一个层级
				return result, i-level - 1
			}else if level > main_level{
				panic(fmt.Sprintf("level is not different,near in '%s...'",buf[i:i+10]))
			}
			is_ele = 1
		case '\r','\n':
		//case '\n':
			//if level!=main_level + 2{
			//	panic(fmt.Sprintf("it's not same level, near in '%s'",string(buf[start:start+10])))
			//}
			//level = 0
		case ' ':
			if is_ele==1 {
				start = skin_space(i+1,buf)
				//if value,end=parse_object2(level,start,0,buf);end==-1{
				//	panic(fmt.Sprintf("cannot get element, near in '%s'",string(buf[start:start+10])))
				//}
				if value,end=parse_value(variable,level + 2,start,0,buf);end==-1{ //level + 2  除了‘-’前面的空格 还包含了”-“自身和后面的一个空格即”- “
					panic(fmt.Sprintf("cannot get element, near in '%s'",string(buf[start:start+10])))
				}
				if var_name,ok:=value.(string);ok{
					if var_name[0]=='&'{
						if value,_ = variable[var_name];value==nil{
							panic("unavailable variable:" + var_name)
						}else{
							result= append(result, value.([]interface{})...)
						}
					}else{
						result = append(result,value)
					}
				}else{
					result = append(result,value)
				}
				i,is_ele,end,level,value = end,0, 0,0, nil
			}else if is_ele == 0{
				level++
			}
		default:
			if level != main_level{
				return result ,i-1-level
			}
		}
	}

	return result,i
}


func skin_white(offset int,buf []byte) int{
	size,i:=len(buf),offset
	for ;i < size;i++{ //空格 或者 tab
		c:= buf[i]
		if c!=' ' && c!='\t' && c!='\r' && c!='\n' && c !='\f'{
			return i
		}
	}
	return size
}

func end_white_bytes(offset int,data []byte) int{
	size,i,is_r:= len(data),offset,0

	for ;i < size;i++{
		c:= data[i]
		if c==' ' ||c=='\n' {
			return i-is_r
		}else if c=='\r' {
			is_r = 1
		}
	}

	return size
}
func skin_space(pos int, data []byte) int{
	size,i:= len(data),pos
	for ;i < size;i++{
		c:= data[i]
		if c!=' '{
			return i
		}
	}

	return size
}

func skin_notes(pos int, data []byte) int{
	size,i:= len(data),pos
	for ;i < size;i++{
		c:= data[i]
		if c=='\n'{
			return i
		}
	}

	return size
}

func get_key(start int,end int, buf []byte) string{
	i:=start
	for ;i< end;i++{
		if buf[i] ==' '{
			end = i
			break
		}
	}

	if end >start{
		return string(buf[start:end])
	}

	return ""
}

func find_index_byte(offset int,data []byte,sep byte) int{
	for i:=offset;i< len(data);i++{
		if data[i]==sep{
			return i
		}
	}
	return -1
}

//reflect.ValueOf(a) 获取的是a的非指针值类型
//reflect.ValueOf(a).Elem() 如果a是指针 则Elem()是把a转成非指针值类型 非指针原型不能再使用.Elem()

func format_object(level int,value *reflect.Value,line bool) ([]byte,error){
	switch value.Kind() {
	case reflect.Int,reflect.Int8, reflect.Int16,reflect.Int32,reflect.Int64:
		return []byte(fmt.Sprintf("%d",value.Int())),nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(fmt.Sprintf("%d",value.Uint())),nil
	case reflect.Float32,reflect.Float64:
		return []byte(strconv.FormatFloat(value.Float(),'f',-1,64)),nil
	case reflect.Bool:
		return []byte(fmt.Sprintf("%v",value.Bool())),nil
	case reflect.String:
		return []byte(fmt.Sprintf("`%s`",value.String())) ,nil
	case reflect.Array://静态数组
		fallthrough
	case reflect.Slice://动态数组(切片)
		return format_list(level,value)
	case reflect.Ptr://指针
		if value.IsNil(){
			return []byte("null"),nil
		}else{
			e:=value.Elem()
			return format_object(level+2, &e,line)
		}
	case reflect.Struct://结构
		if level>0{
			if data,err:=format_struct(level,value);err!=nil{
				return nil,errors.New("marshal struct fail")
			}else{
				if line{ //换行
					return append([]byte{'\n'},data...),nil
				}else{
					return data,nil
				}

			}
		}
		return format_struct(level+2,value)
	case reflect.Map://键值
		return format_map(level,value)
	case reflect.Invalid://空值
		return []byte("null"),nil
	case reflect.Interface:
		e:=value.Elem()
		return format_object(level,&e,true)
	}

	return nil,nil
}

func format_map(level int,value *reflect.Value) ([]byte,error){
	buf:=bytes.Buffer{}
	valueOfKeys:=value.MapKeys()
	if size:=len(valueOfKeys);size>0{
		buf.WriteByte('\n')
		for i:=0;i<size ;i++{
			if valueOfKeys[i].Kind()!=reflect.String{
				return nil,errors.New("marshal map fail,key must be a string type")
			}
			k:=valueOfKeys[i].String()
			if k[0]=='&'{
				continue
			}
			buf.WriteString(fmt.Sprintf("%s%s:",strings.Repeat(" ",level),k))

			v:=value.MapIndex(valueOfKeys[i])
			if data,err := format_object(level*2,&v,true);err!=nil{
				return nil,errors.New("marshal map fail")
			}else{
				if data!=nil{
					buf.Write(data)
					if i<size-1{
						buf.WriteString("\n")
					}
				}
			}
		}
	}else{
		buf.WriteString("null")
	}
	return buf.Bytes(),nil
}

func format_list(level int,value *reflect.Value) ([]byte,error){
	size:=value.Len()
	buf:=bytes.Buffer{}
	if size>0{
		buf.WriteByte('\n')
		for i:=0;i<size;i++{
			valueOf:=value.Index(i)
			if data,err:=format_object(level*2, &valueOf,false);err!=nil{
				return nil,errors.New("marshal list fail")
			}else{
				l:=level
				if l==0{
					l=2
				}
				buf.WriteString(fmt.Sprintf("%s- ",strings.Repeat(" ",l)))
				if level*2< len(data){
					buf.Write(data[level*2:])
				}else{
					buf.Write(data)
				}
				if i<size-1{
					buf.WriteByte('\n')
				}
			}
		}
	}else{
		buf.WriteString("[]")
	}

	return buf.Bytes(),nil
}

func format_struct(level int ,value *reflect.Value) ([]byte,error) {
	buf:=bytes.Buffer{}
	size:=value.NumField()
	for i:=0;i<size;i++ {
		field := value.Type().Field(i)
		tag := field.Tag
		key := tag.Get("json")

		if len(key) == 0 {
			key = field.Name
		}else if key[0]=='-'{
			continue
		}else if key[0]=='&'{
			continue
		}
		buf.WriteString(fmt.Sprintf("%s%s:",strings.Repeat(" ",level),key))

		valueOf := value.FieldByName(field.Name)
		if data,err := format_object((level+1)*2,&valueOf,true);err!=nil{
			return nil,errors.New("marshal struct fail")
		}else{
			buf.Write(data)
			if i<size-1{
				buf.WriteString("\n")
			}
		}
	}
	return buf.Bytes(),nil
}


func set_object(value *reflect.Value,object interface{}) error{
	switch value.Kind() {
	case reflect.Int, reflect.Int8,reflect.Int16,reflect.Int32, reflect.Int64:
		o:=reflect.ValueOf(object)
		if o.Type().Kind()>=reflect.Uint && o.Type().Kind()<=reflect.Uint64{
			if v,ok:=object.(uint64);ok{
				value.SetInt(int64(v))
			}
		}else {
			if v,ok:=object.(int64);ok{
				value.SetInt(v)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		o:=reflect.ValueOf(object)
		if o.Type().Kind()>=reflect.Int && o.Type().Kind()<=reflect.Int64{
			if v,ok:=object.(int64);ok{
				value.SetUint(uint64(v))
			}
		}else {
			if v,ok:=object.(uint64);ok{
				value.SetUint(v)
			}
		}
	case reflect.Float32,reflect.Float64:
		o:=reflect.ValueOf(object)
		if o.Type().Kind()>=reflect.Uint && o.Type().Kind()<=reflect.Uint64{
			if v,ok:=object.(uint64);ok{
				value.SetFloat(float64(v))
			}
		}else if o.Type().Kind()>=reflect.Int && o.Type().Kind()<=reflect.Int64{
			if v,ok:=object.(int64);ok{
				value.SetFloat(float64(v))
			}
		}else{
			if v,ok:=object.(float64);ok{
				value.SetFloat(v)
			}
		}
	case reflect.Bool:
		if v,ok:=object.(bool);ok{
			if v{
				value.SetBool(true)
			}else{
				value.SetBool(false)
			}
		}
	case reflect.String:
		if v,ok:=object.(string);ok{
			value.SetString(v)
		}
	case reflect.Array://静态数组
		fallthrough
	case reflect.Slice://动态数组(切片)
		set_list(value,object.([]interface{}))
	case reflect.Ptr://指针
		if value.IsNil(){
			value.Set(reflect.New(value.Type().Elem()))
			return set_object(value,object)
		}else{
			e:=value.Elem()
			return set_object(&e,object)
		}
	case reflect.Struct://结构
		if object!=nil{
			set_struct(value,object.(map[string]interface{}))
		}
	case reflect.Map://键值
		if object!=nil{
			set_map(value,object.(map[string]interface{}))
		}
	case reflect.Interface:
		value.Set(reflect.ValueOf(object))
	default:
		return errors.New(errors.New("unknown value kind:"+value.Kind().String()).Error())
	}
	return nil
}


func set_struct(value *reflect.Value,object map[string]interface{}) error{
	size:=value.NumField()
	for i:=0;i<size;i++ {
		field := value.Type().Field(i)
		tag := field.Tag
		key := tag.Get("json")

		if len(key) == 0 {
			key = field.Name
		}else if key[0]=='-'{
			continue
		}

		if o,ok:=object[key];ok{
			valueOf := value.FieldByName(field.Name)
			set_object(&valueOf,o)
		}
	}

	return nil
}

func set_list(value *reflect.Value,object []interface{}) error{
	size:=len(object)
	if size>0{
		var v reflect.Value
		slice:=reflect.MakeSlice(value.Type(),size,size)
		for i:=0;i<size;i++{
			v = slice.Index(i)
			set_object(&v,object[i])
		}
		value.Set(slice)
	}
	return nil
}

func set_map(value *reflect.Value,object map[string]interface{}) error{
	var ro reflect.Value
	vm := reflect.MakeMap(value.Type())

	for k, o := range object {
		rk:=reflect.ValueOf(k)
		if value.Type().Elem().Kind()==reflect.Interface{
			ro = reflect.ValueOf(o)
		}else{
			ro = reflect.New(value.Type().Elem()).Elem()
			if err:=set_object(&ro,o);err!=nil{
				panic(fmt.Sprintf("%s=%+v %s",k,o,err.Error()))
			}
		}
		vm.SetMapIndex(rk,ro)
	}

	value.Set(vm)
	return nil
}
