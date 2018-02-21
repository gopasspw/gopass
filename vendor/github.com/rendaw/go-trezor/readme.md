# Trezor
### In Go

This is a direct port of [python-trezor](https://github.com/trezor/python-trezor)
with minimal source code changes for language differences.

The `Client` class has been omitted - you'll have to marshal and loop on
the responses yourself.

As an example, this is from [go-trezor-encrypt](https://github.com/rendaw/go-trezor-encrypt):

```
t := true
err := device.Write(&messages.CipherKeyValue{
    AddressN:     []uint32{10, 0},
    Key:          &key,
    Value:        padded[:],
    Encrypt:      &encrypt,
    AskOnEncrypt: &t,
    AskOnDecrypt: &t,
    Iv:           []byte{},
})
if err != nil {
    return nil, err
}
for {
    messageType, message, err := device.Read()
    if err != nil {
        return nil, err
    }
    if false {

    } else if messageType == messages.MessageType_MessageType_Failure {
        var failure messages.Failure
        proto.Unmarshal(message, &failure)
        return nil, fmt.Errorf("[%d] %s", failure.Code, *failure.Message)
    } else if messageType == messages.MessageType_MessageType_PinMatrixRequest {
        done, value, err := DoPin(prompt)
        if err != nil {
            return nil, err
        }
        if done {
            device.Write(&messages.PinMatrixAck{Pin: &value})
        } else {
            device.Write(&messages.Cancel{})
            return nil, nil
        }
    } else if messageType == messages.MessageType_MessageType_ButtonRequest {
        device.Write(&messages.ButtonAck{})
    } else if messageType == messages.MessageType_MessageType_CipheredKeyValue {
        var ciphered messages.CipheredKeyValue
        proto.Unmarshal(message, &ciphered)
        var clipped []byte
        if encrypt {
            clipped = ciphered.Value
        } else {
            length := binary.BigEndian.Uint32(ciphered.Value[0:4])
            clipped = ciphered.Value[4 : 4+length]
        }
        return clipped, nil
    } else {
        return nil, fmt.Errorf("Unexpected response from trezor [%d]", messageType)
    }
}
```

Currently only protocol V1 and HID V2 have been tested.