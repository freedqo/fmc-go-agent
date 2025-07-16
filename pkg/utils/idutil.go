package utils

import (
	"github.com/sony/sonyflake"
	"net"
	"strconv"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	st.MachineID = func() (uint16, error) {
		ip := GetLocalIP()
		return uint16([]byte(ip)[2])<<8 + uint16([]byte(ip)[3]), nil
	}
	sf = sonyflake.NewSonyflake(st)
}

// GetIntID returns uint64 uniq id.
func GetIntID() uint64 {
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}

	return id
}
func GetStringID() string {
	return strconv.FormatUint(GetIntID(), 10)
}

// GetLocalIP returns the non loopback local IP of the host.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
