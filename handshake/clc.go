package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

var (
	// pcap variables
	pcapDevice  = flag.String("i", "eth0", "the interface to listen on")
	pcapPromisc = flag.Bool("promisc", true, "promiscuous mode")

	// smc variables
	smcrOption = []byte{0xE2, 0xD4, 0xC3, 0xD9}

	// flow table
	flows = make(map[gopacket.Flow]map[gopacket.Flow]bool)
)

const clcHeaderLen = 8

type clcHeader struct { /* header1 of clc messages */
	eyecatcher [4]byte
	typ        uint8 /* proposal / accept / confirm / decline */
	length     uint16
	flags      uint8
}

// smcStreamFactory implementing tcpassembly.StreamFactory
type smcStreamFactory struct{}

// smcStream for decoding of smc packets
type smcStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

// create new smc stream factory (-> implement tcpassembly.StreamFactory)
func (h *smcStreamFactory) New(
	net, transport gopacket.Flow) tcpassembly.Stream {
	sstream := &smcStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	go sstream.run() // parse stream in goroutine

	// ReaderStream implements tcpassembly.Stream, so we can return a
	// pointer to it.
	return &sstream.r
}

// parse smc stream
func (s *smcStream) run() {
	buf := make([]byte, 2048)
	smc := clcHeader{}
	total := 0

	for {
		n, err := s.r.Read(buf[total:])
		total += n
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error reading stream:", err)
		} else {
			if total < clcHeaderLen {
				continue
			}

			if bytes.Compare(buf[0:4], smcrOption) != 0 {
				break
			}
			copy(smc.eyecatcher[:], buf[0:4])
			smc.typ = buf[4]
			smc.length = binary.BigEndian.Uint16(buf[5:7])
			smc.flags = buf[7]
			break
		}
	}
	fmt.Println("SMC flow:           ", s.net, s.transport)
	fmt.Println("With CLC Header:    ", smc.typ, smc.flags)
	tcpreader.DiscardBytesToEOF(&s.r)
}

// check if SMC option is set in TCP header
func checkSMCOption(tcp *layers.TCP) bool {
	for _, opt := range tcp.Options {
		if opt.OptionType == 254 &&
			opt.OptionLength == 6 &&
			bytes.Compare(opt.OptionData, smcrOption) == 0 {
			return true
		}
	}

	return false
}

// listen on network interface and parse packets
func listen() {
	pcapSnaplen := int32(1024)

	// open device
	pcapHandle, pcapErr := pcap.OpenLive(*pcapDevice, pcapSnaplen,
		*pcapPromisc, pcap.BlockForever)
	if pcapErr != nil {
		log.Fatal(pcapErr)
	}
	defer pcapHandle.Close()

	// Set up assembly
	streamFactory := &smcStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(pcapHandle,
		pcapHandle.LinkType())
	for packet := range packetSource.Packets() {
		// only handle tcp packets (with valid network layer)
		if packet.NetworkLayer() == nil ||
			packet.TransportLayer() == nil ||
			packet.TransportLayer().LayerType() !=
				layers.LayerTypeTCP {
			continue
		}
		tcp, ok := packet.TransportLayer().(*layers.TCP)
		if !ok {
			log.Fatal("Error parsing TCP packet")
		}

		// if smc option is set, try to parse tcp stream
		nflow := packet.NetworkLayer().NetworkFlow()
		tflow := packet.TransportLayer().TransportFlow()
		if checkSMCOption(tcp) || flows[nflow][tflow] {
			if flows[nflow] == nil {
				flows[nflow] = make(map[gopacket.Flow]bool)
			}
			flows[nflow][tflow] = true
			assembler.AssembleWithTimestamp(nflow, tcp,
				packet.Metadata().Timestamp)
		}
	}
}

// main
func main() {
	flag.Parse()
	listen()
}
