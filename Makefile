BINDIR := bin
CMDS := memlat runqlat iolat

.PHONY: all build clean $(CMDS) help

all: build

help:
	@echo "Targets:"
	@echo "  make build      - build all Go CLIs into ./bin"
	@echo "  make memlat     - build memlat only"
	@echo "  make runqlat    - build runqlat only"
	@echo "  make iolat      - build iolat only"
	@echo "  make clean      - remove ./bin"

build: $(CMDS)

$(BINDIR):
	mkdir -p $(BINDIR)

memlat: | $(BINDIR)
	go build -o $(BINDIR)/memlat ./cmd/memlat

runqlat: | $(BINDIR)
	go build -o $(BINDIR)/runqlat ./cmd/runqlat

iolat: | $(BINDIR)
	go build -o $(BINDIR)/iolat ./cmd/iolat

clean:
	rm -rf $(BINDIR)

BPF_CLANG ?= clang
BPF_CFLAGS := -O2 -g -target bpf -D__TARGET_ARCH_x86

bpf-runqlat:
	$(BPF_CLANG) $(BPF_CFLAGS) -I./bpf -c bpf/runqlat.bpf.c -o bpf/runqlat.bpf.o
