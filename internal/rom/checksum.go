package rom

import "io"

// UpdateChecksum mirrors app/Rom.php:127-152.
// LoROM: sum all bytes except 0x7FDC-0x7FDF, store inverse at 0x7FDC and checksum at 0x7FDE.
func (r *ROM) UpdateChecksum() error {
	if _, err := r.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	sum := 0x1FE
	buf := make([]byte, 1024)
	for blockStart := 0; blockStart < Size; blockStart += 1024 {
		if _, err := io.ReadFull(r.f, buf); err != nil {
			return err
		}
		for j := range 1024 {
			abs := blockStart + j
			if abs >= 0x7FDC && abs < 0x7FE0 {
				continue
			}
			sum += int(buf[j])
		}
	}

	checksum := sum & 0xFFFF
	inverse := checksum ^ 0xFFFF

	out := []byte{
		byte(inverse & 0xFF), byte((inverse >> 8) & 0xFF),
		byte(checksum & 0xFF), byte((checksum >> 8) & 0xFF),
	}
	return r.Write(0x7FDC, out, true)
}
