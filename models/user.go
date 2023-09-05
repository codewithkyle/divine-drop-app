package models

import (
    "bytes"
    "encoding/gob"
    "log"
)

type User struct {
    Id string
    Username string
    Email string
    Avatar string
}


func UserToBlob(data User) ([]byte, error) {
    var buffer bytes.Buffer
    encoder := gob.NewEncoder(&buffer)
    err := encoder.Encode(data)
    if err != nil {
        log.Fatal("encode error:", err)
        return nil, err
    }
    return buffer.Bytes(), nil
}

func BlobToUser(blobData []byte) (User, error) {
    var data User
    buffer := bytes.NewBuffer(blobData)
    decoder := gob.NewDecoder(buffer)
    err := decoder.Decode(&data)
    if err != nil {
        log.Fatal("decode error:", err)
        return User{}, err
    }
    return data, nil
}
