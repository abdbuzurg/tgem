// Package handlers — invoice_document.go
//
// Shared resolver for /document/:deliveryCode endpoints across every invoice
// flavor. Looks up the requested file in the primary location first
// (./storage/import_excel/<flavor>/), then in the legacy bind-mount
// (/app/legacy_excels/<flavor>/) as a read-only fallback for files generated
// before the rewrite cutover. See MIGRATION/06-xlsx-investigation.md.
package handlers

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

// legacyExcelsRoot is the in-container mount point of the legacy host
// directory. Aligned with the bind-mount declared in docker-compose.yml.
// Kept as a var (not const) so tests can rebind it; do not mutate at runtime.
var legacyExcelsRoot = "/app/legacy_excels"

// resolveInvoiceDoc returns the first path that exists among:
//
//	primaryDir/deliveryCode+ext
//	legacyExcelsRoot/legacySubdir/deliveryCode+ext
//
// It returns "" if neither exists. The boolean indicates whether the resolved
// path came from the legacy fallback (for logging only — never expose to the
// client).
func resolveInvoiceDoc(deliveryCode, ext, primaryDir, legacySubdir string) (string, bool) {
	primary := filepath.Join(primaryDir, deliveryCode+ext)
	if _, err := os.Stat(primary); err == nil {
		return primary, false
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Printf("invoice_document: stat %s: %v", primary, err)
	}

	if legacySubdir == "" {
		return "", false
	}
	legacy := filepath.Join(legacyExcelsRoot, legacySubdir, deliveryCode+ext)
	if _, err := os.Stat(legacy); err == nil {
		log.Printf("invoice_document: serving %s from legacy fallback", deliveryCode+ext)
		return legacy, true
	}
	return "", false
}

// resolveInvoiceDocGlob handles the writeoff case where the served file's
// extension is unknown ahead of time and must be discovered via Glob. Same
// primary-then-legacy lookup order as resolveInvoiceDoc.
func resolveInvoiceDocGlob(deliveryCode, primaryDir, legacySubdir string) (string, bool) {
	dirs := []string{primaryDir}
	if legacySubdir != "" {
		dirs = append(dirs, filepath.Join(legacyExcelsRoot, legacySubdir))
	}
	for i, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join(dir, deliveryCode+".*"))
		if err != nil || len(matches) == 0 {
			continue
		}
		if i == 1 {
			log.Printf("invoice_document: serving %s from legacy fallback", filepath.Base(matches[0]))
		}
		return matches[0], i == 1
	}
	return "", false
}
