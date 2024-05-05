package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	// PowerShellコマンドを構築
	cmd := exec.Command("powershell", "Get-NetAdapter -Name '携帯電話' | Get-NetAdapterStatistics | Select-Object Name, ReceivedBytes, SentBytes | Format-Table -HideTableHeaders")

	// コマンドの出力をバッファに
	var out bytes.Buffer
	cmd.Stdout = &out

	// コマンド実行
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 出力の解析
	result := strings.Fields(out.String())
	if len(result) >= 3 {
		fmt.Println("Adapter Name:", result[0])
		fmt.Println("Received Bytes:", result[1])
		fmt.Println("Sent Bytes:", result[2])
	}
}
