package main

import (
    "context"
    "fmt"
    "math/rand"
    "os"
    "strings"
    "time"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"

    _ "github.com/mattn/go-sqlite3"
)

func main() {
    NewBot("6285796103714", func(k string) { //masukkan nomor kamu yang ingin di pasangkan auto read story wa
        println(k)
    }) 
}

func registerHandler(client *whatsmeow.Client) func(evt interface{}) {
    return func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            if v.Info.Chat.String() == "status@broadcast" {
                if v.Info.Type != "reaction" {
                    sender := v.Info.Sender.String()
                    allowedSenders := []string{ //disini isi nomer yang ingin agar bot tidak otomatis read sw dari list nomor dibawah 
                        "6281447477366@s.whatsapp.net",
                        "6281457229553@s.whatsapp.net",
                    }
                    if contains(allowedSenders, sender) {
                        return
                    }

                    emojis := []string{"🔥", "✨", "🌟", "🌞", "🎉", "🎊", "😺"}
                    rand.Seed(time.Now().UnixNano())
                    randomEmoji := emojis[rand.Intn(len(emojis))]

                    reaction := client.BuildReaction(v.Info.Chat, v.Info.Sender, v.Info.ID, randomEmoji)
                    extras := []whatsmeow.SendRequestExtra{}
                    client.MarkRead([]types.MessageID{v.Info.ID}, v.Info.Timestamp, v.Info.Chat, v.Info.Sender)
                    client.SendMessage(context.Background(), v.Info.Chat, reaction, extras...)
                    fmt.Println("Berhasil melihat status", v.Info.PushName)
                }
            }
        }
    }
}

func NewBot(id string, callback func(string)) *whatsmeow.Client {
    if id == "" {
        callback("Nomor ?")
        return nil
    }
    id = strings.ReplaceAll(id, "admin", "")

    dbLog := waLog.Stdout("Database", "ERROR", true)

    container, err := sqlstore.New("sqlite3", "file:"+id+".db?_foreign_keys=on", dbLog)
    if err != nil {
        callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
        return nil
    }
    deviceStore, err := container.GetFirstDevice()
    if err != nil {
        callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
        return nil
    }
    clientLog := waLog.Stdout("Client", "ERROR", true)
    client := whatsmeow.NewClient(deviceStore, clientLog)
    client.AddEventHandler(registerHandler(client))

    err = client.Connect()
    if err != nil {
        callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
        return nil
    }

    if client.Store.ID == nil {
        code, _ := client.PairPhone(id, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
        callback("Kode verifikasi anda adalah " + code)
        time.AfterFunc(60*time.Second, func() {
            if client.Store.ID == nil {
                client.Disconnect()
                os.Remove(id + ".db")
                callback("melebihi 60 detik, memutuskan")
            }
        })
    } else {
        fmt.Println("Connected to readsw!!")
    }
    return client
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
