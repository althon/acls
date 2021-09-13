# Config language script
Config is a file configuration loader for golang
# Features
 * Support string、paragraph、int-int64、uint-uint64、float64、bool、map、array
 * Can be converted to json string
 * Case sensitive
 * Support comments lines
 * Support reference block
# config file example
```
#example 1
username:althon
age:18
```
````
#example 2
list1:
 - swimming
 - basketball
 - more
list2:
 - "this is a double quoted strings"
 - 'you can also use single quotes as strings'
 - `if there are spaces in the string, 
the string must be enclosed in quotation marks`
````
````
#example 3
array1:[1,2,3]
array2:[
1,2,
3
]
array3:[althon,2,3,true,"good boy",'bad boy',`this is a good job`]
````
````
#example 4
user:
  name:althon
  age:100
list:
  - name:althon
    age:10
  - name:jack
    age:100
list2:
   - student1:
      - name:althon
        age:10
      - name:jack
        age:18
   - student2:
      - name:althon
        age:10
      - name:jack
        age:18
paragraph: `Dear:
   Do you know?
   This is a good job`
````
# Code Example
```go
func main(){
    acls.FormFile(file)//Load from file to acls object or
    acls.ToAcls([]byte)//Load from bytes to acls object or
    acls.Marshal(interface{})//Marshal from interface to bytes or 
    acls.ToJson([]byte acls)//load from acls bytes to json bytes
    acls.Unmarshal([]byte,&interface{}) //Unmarshal from bytes to interface{}
}
```
```go
func main(){
    cls:=acls.FormFile(`file`)//Load from file to acls object or
	
  //for example #4
    fmt.Println(cls.Value("user.name")) //result: althon or
    fmt.Println(cls.String("user.name")) //result: althon 
    fmt.Println(cls.Value("user.age")) //result: 10 or 
    fmt.Println(cls.Int("user.age")) //result: 10

    fmt.Println(cls.String("list[0].name")) //result: althon
    fmt.Println(cls.Value("list[1].name")) //result: jack

    fmt.Println(cls.Value("paragraph")) //result: Dear:......
}
```

#'/' of strings usage example
````
#config file
#use  
text1:`  /
This is a description document. Please pay attention to the following points:
1.test test test /
test2 test3 test4
2.day day day /
"day" 'day'
`
text2:"This is a description document. Please pay attention to the following points:
1.test test test /
test2 test3 test4
2.day day day /
\"day\" 'day'"
text3:'This is a description document. Please pay attention to the following points:
1.test test test /
test2 test3 test4
2.day day day /
"day" \'day\''
````
````go
func main() {
    o:=acls.FromFile(config file)
    
    fmt.Println(o.Value("text1"))
    fmt.Println(o.Value("text2"))
    fmt.Println(o.Value("text3"))
}
````
````
result:
  This is a description document. Please pay attention to the following points:
1.test test test test2 test3 test4
2.day day day "day" 'day'
  This is a description document. Please pay attention to the following points:
1.test test test test2 test3 test4
2.day day day "day" 'day'
   This is a description document. Please pay attention to the following points:
1.test test test test2 test3 test4
2.day day day "day" 'day'
	
````
# '$variable','&variable' Reference block
````
#config file 
test1: $a "this is a test"
test2: $b 123
test3: $c
  - 1
  - 2
  - true
test4: $d
  name: althon
  age: 18
  money:&b
test5: &a
test6:&c
test7:&d
````
````go
func main() {
    o:=acls.FromFile(config file)
    
    fmt.Println(o.Value("&a"))
    fmt.Println(o.Value("test4"))
    fmt.Println(o.Value("test6"))
}
````
````
result:
this is a test
map[name:althon age:18 money:123]
[1 2 true]
````