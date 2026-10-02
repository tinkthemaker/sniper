package ui

import "sniper/internal/store"

// ShowRecovery is a placeholder until M6: it shows why the library could not be
// opened and changes nothing on disk.
func ShowRecovery(d Deps, res store.LoadResult) {
	ShowFatal(d.App, "Sniper could not open your snippet library.", res.Err.Error()+"\n\nYour data has not been changed.")
}
