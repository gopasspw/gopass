package trezor

import (
	"fmt"
	"os"
	"reflect"

	"github.com/rendaw/go-trezor/messages"

	"github.com/golang/protobuf/proto"
	"github.com/karalabe/hid"
)

func DevTrezor1() (uint16, uint16)   { return 0x534c, 0x0001 }
func DevTrezor2() (uint16, uint16)   { return 0x1209, 0x53c1 }
func DevTrezor2BL() (uint16, uint16) { return 0x1209, 0x53c0 }

func IsWirelink(dev *hid.DeviceInfo) bool {
	return dev.UsagePage == 0xFF00 || dev.Interface == 0
}

func IsDebuglink(dev *hid.DeviceInfo) bool {
	return dev.UsagePage == 0xFF01 || dev.Interface == 1
}

func IsTrezor1(dev *hid.DeviceInfo) bool {
	gotVendor, gotProduct := DevTrezor1()
	return dev.VendorID == gotVendor && dev.ProductID == gotProduct
}
func IsTrezor2(dev *hid.DeviceInfo) bool {
	gotVendor, gotProduct := DevTrezor2()
	return dev.VendorID == gotVendor && dev.ProductID == gotProduct
}
func IsTrezor2BL(dev *hid.DeviceInfo) bool {
	gotVendor, gotProduct := DevTrezor2BL()
	return dev.VendorID == gotVendor && dev.ProductID == gotProduct
}

func Enumerate() ([]*HidTransport, error) {
	var out []*HidTransport
	for _, dev := range hid.Enumerate(0, 0) {
		if !(IsTrezor1(&dev) || IsTrezor2(&dev) || IsTrezor2BL(&dev)) {
			continue
		}
		if !IsWirelink(&dev) {
			continue
		}
		transport, err := HidTransportNew(dev)
		if err != nil {
			return nil, err
		}
		out = append(out, transport)
	}
	return out, nil
}

type HidHandle struct {
	count  int
	Handle *hid.Device
}

func (self *HidHandle) Open(info *hid.DeviceInfo) error {
	if self.count == 0 {
		handle, err := info.Open()
		if err != nil {
			return err
		}
		self.Handle = handle
	}
	self.count += 1
	return nil
}

func (self *HidHandle) Close() error {
	if self.count == 1 {
		err := self.Handle.Close()
		if err != nil {
			return err
		}
	}
	if self.count > 0 {
		self.count -= 1
	}
	return nil
}

type HidTransport struct {
	Info       hid.DeviceInfo
	Hid        HidHandle
	HidVersion int
	protocol   Protocol
}

func HidTransportNew(info hid.DeviceInfo) (*HidTransport, error) {
	forceV1, found := os.LookupEnv("TREZOR_TRANSPORT_V1")
	var protocol Protocol
	if IsTrezor2(&info) || found && forceV1 != "1" {
		var err error
		protocol, err = ProtocolV2New()
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		protocol, err = ProtocolV1New()
		if err != nil {
			return nil, err
		}
	}
	return &HidTransport{
		Info: info,
		Hid: HidHandle{
			count:  0,
			Handle: nil,
		},
		protocol: protocol,
	}, nil
}

func (self *HidTransport) Open() error {
	err := self.Hid.Open(&self.Info)
	if err != nil {
		return fmt.Errorf("Unable to open device %s: %s", self.Info.Path, err)
	}
	if IsTrezor1(&self.Info) {
		if self.HidVersion, err = ProbeHidVersion(self); err != nil {
			return err
		}
	} else {
		self.HidVersion = 2
	}
	self.protocol.SessionBegin(self)
	return nil
}

func ProbeHidVersion(self *HidTransport) (int, error) {
	data := [65]byte{}
	data[0] = 0
	data[1] = 63
	for i := 2; i < len(data); i++ {
		data[i] = 0xFF
	}
	{
		n, err := self.Hid.Handle.Write(data[:])
		if err != nil {
			return 0, err
		}
		if n == 65 {
			return 2, nil
		}
	}
	{
		n, err := self.Hid.Handle.Write(data[1:])
		if err != nil {
			return 0, err
		}
		if n == 64 {
			return 1, nil
		}
	}
	return 0, fmt.Errorf("Unknown HID version")
}

func (self *HidTransport) Close() error {
	self.protocol.SessionEnd(self)
	err := self.Hid.Close()
	if err != nil {
		return err
	}
	self.HidVersion = -1
	return nil
}

func (self *HidTransport) Read() (messages.MessageType, []byte, error) {
	return self.protocol.Read(self)
}

func (self *HidTransport) ReadChunk() ([]byte, error) {
	chunk := [64]byte{}
	for {
		read, err := self.Hid.Handle.Read(chunk[:])
		if err != nil {
			return nil, err
		}
		if read != 64 {
			return nil, fmt.Errorf("Unexpected chunk size %d", read)
		}
		break
	}
	return chunk[:], nil
}

func (self *HidTransport) Write(message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	typeName := "MessageType_" + reflect.TypeOf(message).Elem().Name()
	messageId, found := messages.MessageType_value[typeName]
	if !found {
		return fmt.Errorf("Could not send message; unknown message type %s", typeName)
	}
	self.protocol.Write(self, messages.MessageType(messageId), data)
	return nil
}

func (self *HidTransport) WriteChunk(chunk []byte) error {
	if len(chunk) != 64 {
		return fmt.Errorf("Unexpected chunk size: %d", len(chunk))
	}
	if self.HidVersion == 2 {
		if _, err := self.Hid.Handle.Write(append([]byte{0}, chunk...)); err != nil {
			return err
		}
	} else {
		if _, err := self.Hid.Handle.Write(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (self *HidTransport) String() string {
	return self.Info.Path
}
