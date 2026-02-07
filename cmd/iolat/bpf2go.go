//go:build linux
package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -D__TARGET_ARCH_x86 -I../../bpf -I/usr/include/bpf" iolat ../../bpf/iolat.bpf.c

// (옵션) 생성된 파일이 go test/build에 포함되도록 아무것도 안 해도 됨.
// 이 파일은 "go generate ./cmd/iolat"를 위해 존재함.
