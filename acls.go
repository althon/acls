package acls

import (
	"strconv"
	"strings"
)

type Acls map[string]interface{}

func (p Acls) Value(key string) interface{}{
	var err error
	keys := strings.Split(key,".")
	acls,index,size := p,-1, len(keys)

	for i:=0;i< size;i++{
		is_array_start:=strings.Index(keys[i],"[")
		if is_array_start!=-1{
			is_array_end:=strings.Index(keys[i],"]")
			if is_array_end!=-1{
				index,err = strconv.Atoi(keys[i][is_array_start+1:is_array_end])
				if err!=nil{
					panic("unavailable keys:" + keys[i])
				}
			}
			keys[i]=keys[i][:is_array_start]
		}

		if a,ok:=acls[keys[i]];ok{
			if index!=-1{ //数组
				index_value:=a.([]interface{})[index]
				if aa,ok:=index_value.(map[string]interface{});ok{
					acls = aa
				}else{
					if i==size-1{
						return index_value
					}
					panic("unavailable keys:" +key)
				}
				index = -1
			}else{
				if aa,ok:=a.(map[string]interface{});ok{
					acls = aa
				}else{
					if i==size-1{
						return a
					}
					panic("unavailable keys:" + key)
				}
			}

		}else{
			panic("unavailable keys:" + key)
		}
	}
	return nil
}


func (p Acls) Int(key string) int{
	return int(p.Value(key).(int64))
}

func (p Acls) Int8(key string) int8{
	return int8(p.Value(key).(int64))
}

func (p Acls) Int16(key string) int16{
	return int16(p.Value(key).(int64))
}

func (p Acls) Int32(key string) int32{
	return int32(p.Value(key).(int64))
}

func (p Acls) Int64(key string) int64{
	return p.Value(key).(int64)
}

func (p Acls) Uint(key string) uint{
	return uint(p.Value(key).(uint64))
}

func (p Acls) Uint8(key string) uint8{
	return uint8(p.Value(key).(uint64))
}

func (p Acls) Uint16(key string) uint16{
	return uint16(p.Value(key).(uint64))
}

func (p Acls) Uint32(key string) uint32{
	return uint32(p.Value(key).(uint64))
}

func (p Acls) Uint64(key string) uint64{
	return p.Value(key).(uint64)
}

func (p Acls) Float32(key string) float32{
	return float32(p.Value(key).(float64))
}

func (p Acls) Float(key string) float64{
	return p.Value(key).(float64)
}

func (p Acls) String(key string) string{
	return p.Value(key).(string)
}

func (p Acls) Bool(key string) bool{
	return p.Value(key).(bool)
}