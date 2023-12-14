// client create: PingerClient
/*
  Created by /home/cnw/devel/go/yatools/src/golang.yacloud.eu/yatools/protoc-gen-cnw/protoc-gen-cnw.go
*/

/* geninfo:
   filename  : protos/golang.conradwood.net/apis/pinger/pinger.proto
   gopackage : golang.conradwood.net/apis/pinger
   importname: ai_0
   clientfunc: GetPinger
   serverfunc: NewPinger
   lookupfunc: PingerLookupID
   varname   : client_PingerClient_0
   clientname: PingerClient
   servername: PingerServer
   gsvcname  : pinger.Pinger
   lockname  : lock_PingerClient_0
   activename: active_PingerClient_0
*/

package pinger

import (
   "sync"
   "golang.conradwood.net/go-easyops/client"
)
var (
  lock_PingerClient_0 sync.Mutex
  client_PingerClient_0 PingerClient
)

func GetPingerClient() PingerClient { 
    if client_PingerClient_0 != nil {
        return client_PingerClient_0
    }

    lock_PingerClient_0.Lock() 
    if client_PingerClient_0 != nil {
       lock_PingerClient_0.Unlock()
       return client_PingerClient_0
    }

    client_PingerClient_0 = NewPingerClient(client.Connect(PingerLookupID()))
    lock_PingerClient_0.Unlock()
    return client_PingerClient_0
}

func PingerLookupID() string { return "pinger.Pinger" } // returns the ID suitable for lookup in the registry. treat as opaque, subject to change.

func init() {
   client.RegisterDependency("pinger.Pinger")
   AddService("pinger.Pinger")
}





