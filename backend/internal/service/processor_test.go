package service

import (
	"testing"

	"github.com/aitjcize/esp32-photoframe-server/backend/pkg/photoframe"
)

// compressDynamicRange must reach the CLI in both states. The CLI's default
// preset (balanced) has compression on, so leaving the flag out when the
// device says false silently keeps it on.
func TestMapProcessingSettings_CompressDynamicRange(t *testing.T) {
	proc := NewProcessorService()

	on := proc.MapProcessingSettings(&photoframe.ProcessingSettings{CompressDynamicRange: true}, nil, false)
	if _, ok := on["compress-dynamic-range"]; !ok {
		t.Errorf("true: expected --compress-dynamic-range, got %v", on)
	}
	if _, ok := on["no-compress-dynamic-range"]; ok {
		t.Errorf("true: unexpected --no-compress-dynamic-range in %v", on)
	}

	off := proc.MapProcessingSettings(&photoframe.ProcessingSettings{CompressDynamicRange: false}, nil, false)
	if _, ok := off["no-compress-dynamic-range"]; !ok {
		t.Errorf("false: expected --no-compress-dynamic-range, got %v", off)
	}
	if _, ok := off["compress-dynamic-range"]; ok {
		t.Errorf("false: unexpected --compress-dynamic-range in %v", off)
	}

	// No settings at all: neither flag, the CLI keeps its defaults.
	none := proc.MapProcessingSettings(nil, nil, false)
	for _, k := range []string{"compress-dynamic-range", "no-compress-dynamic-range"} {
		if _, ok := none[k]; ok {
			t.Errorf("nil settings: unexpected --%s in %v", k, none)
		}
	}
}

// A panel that dithers on-device must get tone-mapped full-color PNG: no
// palette, no dithering, no compression to a measured palette -- even when
// the stored settings ask for them -- while tone and layout options survive.
func TestApplyOnDeviceDithering(t *testing.T) {
	proc := NewProcessorService()
	palette := &photoframe.Palette{White: photoframe.PaletteColor{R: 200, G: 200, B: 190}}
	opts := proc.MapProcessingSettings(&photoframe.ProcessingSettings{
		Exposure:             1.2,
		Saturation:           1.1,
		Contrast:             1.3,
		ToneMode:             "contrast",
		ColorMethod:          "lab",
		DitherAlgorithm:      "stucki",
		CompressDynamicRange: true,
	}, palette, false)
	opts["dimension"] = "2560x1440"
	opts["format"] = "epdgz"

	ApplyOnDeviceDithering(opts)

	for _, k := range []string{"palette", "palette-preset", "dither-algorithm", "color-method", "compress-dynamic-range"} {
		if _, ok := opts[k]; ok {
			t.Errorf("unexpected --%s in %v", k, opts)
		}
	}
	for _, k := range []string{"skip-dithering", "no-compress-dynamic-range"} {
		if v, ok := opts[k]; !ok || v != "" {
			t.Errorf("expected boolean --%s, got %v", k, opts)
		}
	}
	if opts["format"] != "png" {
		t.Errorf("format = %q, want png", opts["format"])
	}
	for k, want := range map[string]string{"exposure": "1.2", "saturation": "1.1", "contrast": "1.3", "tone-mode": "contrast", "dimension": "2560x1440"} {
		if opts[k] != want {
			t.Errorf("--%s = %q, want %q", k, opts[k], want)
		}
	}
}
