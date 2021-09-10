# Config language script
Config is a file configuration loader for golang
# Features
 * Support string、paragraph、int-int64、uint-uint64、float64、bool、map、array
 * Can be converted to json string
 * Case sensitive
 * Support comments lines
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