package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func main() {
	// 設定ファイルを読み込む
	config, err := readConfig("setting.conf")
	if err != nil {
		fmt.Println("Failed to read config:", err)
		return
	}

	// PowerShellコマンドを実行
	cmd := exec.Command("powershell", fmt.Sprintf("Get-NetAdapter -Name '%s' | Get-NetAdapterStatistics | Select-Object Name, ReceivedBytes, SentBytes | Format-Table -HideTableHeaders", config["adapterName"]))
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}

	// 出力の解析
	result := strings.Fields(out.String())
	if len(result) < 3 {
		fmt.Println("Unexpected output format")
		return
	}

	receivedBytes, err := strconv.Atoi(result[1])
	if err != nil {
		fmt.Println("Error parsing received bytes:", err)
		return
	}

	sentBytes, err := strconv.Atoi(result[2])
	if err != nil {
		fmt.Println("Error parsing sent bytes:", err)
		return
	}

	// Prometheusにメトリクスをプッシュ
	if err := pushMetrics(config["pushgatewayAddress"], "network_stats", result[0], receivedBytes, sentBytes); err != nil {
		fmt.Println("Error pushing metrics:", err)
	}
}

func readConfig(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return config, scanner.Err()
}

func pushMetrics(address, jobName, adapterName string, received, sent int) error {
	reg := prometheus.NewRegistry()
	receivedGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "network_received_bytes",
		Help: "Number of bytes received by the adapter",
	})
	sentGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "network_sent_bytes",
		Help: "Number of bytes sent by the adapter",
	})

	receivedGauge.Set(float64(received))
	sentGauge.Set(float64(sent))

	reg.MustRegister(receivedGauge)
	reg.MustRegister(sentGauge)

	return push.New(address, jobName).
		Collector(receivedGauge).
		Collector(sentGauge).
		Grouping("adapter", adapterName).
		Push()
}
