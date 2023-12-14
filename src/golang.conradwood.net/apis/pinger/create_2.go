// client create: PingerListClient
/*
  Created by /home/cnw/devel/go/yatools/src/golang.yacloud.eu/yatools/protoc-gen-cnw/protoc-gen-cnw.go
*/

/* geninfo:
   filename  : protos/golang.conradwood.net/apis/pinger/pinger.proto
   gopackage : golang.conradwood.net/apis/pinger
   importname: ai_1
   clientfunc: GetPingerList
   serverfunc: NewPingerList
   lookupfunc: PingerListLookupID
   varname   : client_PingerListClient_1
   clientname: PingerListClient
   servername: PingerListServer
   gsvcname  : pinger.PingerList
   lockname  : lock_PingerListClient_1
   activename: active_PingerListClient_1
*/

package pinger

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_PingerListClient_1 sync.Mutex
  client_PingerListClient_1 PingerListClient
)

func GetPingerListClient() PingerListClient { 
    if client_PingerListClient_1 != nil {
        return client_PingerListClient_1
    }

    lock_PingerListClient_1.Lock() 
    if client_PingerListClient_1 != nil {
       lock_PingerListClient_1.Unlock()
       return client_PingerListClient_1
    }

    client_PingerListClient_1 = NewPingerListClient(client.Connect(PingerListLookupID()))
    lock_PingerListClient_1.Unlock()
    return client_PingerListClient_1
}

func PingerListLookupID() string { return "pinger.PingerList" } // returns the ID suitable for lookup in the registry. treat as opaque, subject to change.

func init() {
   client.RegisterDependency("pinger.PingerList")
   AddService("pinger.PingerList")
}





