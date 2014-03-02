package cache

import (
    "crypto/md5"
   "encoding/hex"
//   "fmt"
   "os"
   "io/ioutil"
   "bytes"
   "log"
   "encoding/gob"
   "time"
   "path"
   "path/filepath"
)

type FileCache struct{
   Data_dir string
}

type cacheData struct{
    Data string
    Life int64
}

func (cache *FileCache)Set(key string,data string,life int64) (suc bool){
//    log.Println("cache set ",key,data)
   cache_path:=cache.genCachePath(key)
   f,err:=os.OpenFile(cache_path,os.O_CREATE|os.O_RDWR,0644)
   defer f.Close()
   if err!=nil{
       p_dir:=path.Dir(cache_path)
       os.MkdirAll(p_dir,0755)
       f,err=os.OpenFile(cache_path,os.O_CREATE|os.O_RDWR,0644)
       defer f.Close()
    }
   var bf bytes.Buffer
   enc:=gob.NewEncoder(&bf)
   enc.Encode(cacheData{data,time.Now().Unix()+life})
   f.Write(bf.Bytes())
   return true
}

func (cache *FileCache)Get(key string)(has bool,data string){
//    log.Println("cache get ",key)
	 cache_path:=cache.genCachePath(key)
	 return cache.getDataByPath(cache_path)
}

func (cache *FileCache)CheckAll(){
  info,err:=os.Stat(cache.Data_dir)
  if err!=nil || !info.IsDir(){
    return
  }
  filepath.Walk(cache.Data_dir,func(file_path string,info os.FileInfo,err error) error{
     if !info.IsDir(){
         cache.getDataByPath(file_path)
      }
      return nil
  })
  
}
func (cache *FileCache)genCachePath(key string) string{
   h:=md5.New()
   h.Write([]byte(key))
 	md5_str:= hex.EncodeToString(h.Sum(nil))
 	file_path:=cache.Data_dir+"/"+string(md5_str[:3])+"/"+md5_str
 	return file_path
}

func (cache *FileCache)getDataByPath(file_path string)(has bool,data string){
	f,err:=os.Open(file_path)
    defer f.Close()
    if err!=nil{
      return
     }
    data_bf,err1:=ioutil.ReadAll(f)
    if err1!=nil{
    	log.Println("read cache file failed:",file_path,err1.Error())
        return
     }
    dec:= gob.NewDecoder(bytes.NewBuffer(data_bf))
    var cache_data cacheData
    err= dec.Decode(&cache_data)
    if err!=nil{
      return
     }
    if (time.Now().Unix()>cache_data.Life){
      os.Remove(file_path)
      return
     }
   return true,cache_data.Data
}
