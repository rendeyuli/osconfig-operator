package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func ApplyConfig(cm *corev1.ConfigMap) {
	//目前硬编码了两种情况，“hostname”和“sysctl”
	//后续考虑优化为: 定义一个 JSON/YAML 格式，Agent 解析并根据“类型”调用不同函数
	if hostname, ok := cm.Data["hostname"]; ok {
		if err := execCommand("chroot", "/host", "hostnamectl", "set-hostname", hostname); err != nil {
			log.Printf("Failed to set hostname: %v", err)
		}
	}

	if sysctlConf, ok := cm.Data["sysctl"]; ok {
		//多个参数以换行分隔
		for _, line := range splitLines(sysctlConf) {
			if err := execCommand("chroot", "/host", "sysctl", "-w", line); err != nil {
				log.Printf("Failed to set sysctl param %s: %v", line, err)
			}
		}
	}
}

func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(output))
	}
	log.Printf("Executed: %s %v", name, args)
	return nil
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		if trimmed := strings.TrimSpace(l); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}
