package cache

type Cache interface{
   Set(key string,val []byte,life int)(suc bool)
   Get(key string)(has bool,data []byte)
   Delete(key string)(suc bool)
   DeleteAll()(suc bool)
   Clean()
}

type Data struct{
    Key string
    Data []byte
    CreateTime int64
    Life int64
}
