package nes

import (
	"encoding/binary"
	"errors"
	"io"
)

const NESFileMagic = 0x1a53454e

type NESFileHeader struct {
	Magic    uint32
	NumPRG   byte
	NumCHR   byte
	Control1 byte
	Control2 byte
	NumRAM   byte
	_        [7]byte
}

func LoadNES(reader io.Reader) (*Cartridge, error) {
	header := NESFileHeader{}
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	if header.Magic != NESFileMagic {
		return nil, errors.New("invalid .nes file")
	}

	mapper := header.Control1>>4 | header.Control2&0xf0
	mirror := header.Control1&1 | (header.Control1>>3)&1<<1
	battery := (header.Control1 >> 1) & 1

	if header.Control1&4 != 0 {
		trainer := [512]byte{}
		if _, err := io.ReadFull(reader, trainer[:]); err != nil {
			return nil, err
		}
	}

	prg := make([]byte, int(header.NumPRG)*16384)
	if _, err := io.ReadFull(reader, prg); err != nil {
		return nil, err
	}

	chr := make([]byte, int(header.NumCHR)*8192)
	if _, err := io.ReadFull(reader, chr); err != nil {
		return nil, err
	}

	if header.NumCHR == 0 {
		chr = make([]byte, 8192)
	}

	return NewCartridge(prg, chr, mapper, mirror, battery), nil
}
