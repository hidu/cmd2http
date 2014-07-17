package serve

import (
  "strings"
  "os"
  "log"
  "gopkg.in/cookieo9/resources-go.v2"
  "io/ioutil"
   "github.com/hidu/goutils"
)

func GetVersion() string{
   return strings.TrimSpace(string(LoadRes("res/version")));
}

func IsFileExists(path string) bool{
  _,err := os.Stat( path )
  return err==nil
}

func LoadRes(path string) []byte{
     res,err:=GetRes(path)
     if(err!=nil){
        return []byte{}
      }
     r,_:=res.Open()
     bf,err:=ioutil.ReadAll(r)
     if(err!=nil){
        log.Println("read res[",path,"] failed",err.Error())
      }
     return bf
}

func GetRes(path string)(resources.Resource,error){
    path=strings.TrimLeft(path,"/")
    res,err:=resources.Find(path)
    if(err!=nil){
      log.Println("load res[",path,"] failed",err.Error())
      return nil,err
     }
     return res,nil
}


func In_array(item string,arr []string) bool{
  for _,a:=range arr{
     if(a==item){
        return true
       }
   }
 return false
}

func GetCacheKey(cmd string,params []string) string{
     return cmd+"|||"+strings.Join(params,"&")
}

func LoadParamValuesFromFile(file_path string)(values []string){
  if !IsFileExists(file_path){
    return
  }
  bf,err:=utils.File_get_contents(file_path)
  if err!=nil{
     return  
  }
  lines:=strings.Split(string(bf),"\n")
  for _,line:=range lines{
     line=strings.TrimSpace(line)
     if (line=="" || line[0]=='#'){
        continue
      }
    values=append(values,line)
  }
  return
}