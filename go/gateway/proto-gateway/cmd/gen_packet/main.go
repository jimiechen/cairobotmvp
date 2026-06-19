package main

import (
	"fmt"
	"os"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	socialpb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"google.golang.org/protobuf/proto"
)

func main() {
	req := &socialpb.UserLoginRequest{Username: "e2e_full_001", Password: "TestPass123!"}
	reqData, _ := proto.Marshal(req)
	packet := &pb.MessagePacket{
		MaxType: 1000, MinType: 1023,
		Extend: map[string]string{"method": "HandleMember", "minType": "1023"},
		Platform: pb.Platform_WEB, Data: reqData,
	}
	data, _ := proto.Marshal(packet)
	os.WriteFile("/tmp/login_1023.bin", data, 0644)
	fmt.Printf("Written %d bytes\n", len(data))
}
