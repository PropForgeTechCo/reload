package reload

import (
	"errors"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"
)

// WatchDirectories listens for changes in directories and
// broadcasts on write.
func (reload *Reloader) WatchDirectories() {
	if len(reload.directories) == 0 {
		return
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		reload.Log.Printf("error initializing fsnotify watcher: %s\n", err)
	}

	defer func(w *fsnotify.Watcher) {
		err := w.Close()
		if err != nil {
			reload.Log.Printf("error closing fsnotify watcher: %s\n", err)
		}
	}(w)

	for _, p := range reload.directories {
		directories, err := recursiveWalk(p)
		if err != nil {
			var pathErr *fs.PathError
			if errors.As(err, &pathErr) {
				reload.Log.Printf("directory doesn't exist: %s\n", pathErr.Path)
			} else {
				reload.Log.Printf("error walking directories: %s\n", err)
			}
			return
		}
		for _, dir := range directories {
			err := w.Add(dir)
			if err != nil {
				return
			}
		}
	}

	db := debounce.New(100 * time.Millisecond)

	callback := func(path string) func() {
		return func() {
			reload.Log.Println("Edit", path)
			if reload.OnReload != nil {
				reload.OnReload()
			}
			reload.cond.Broadcast()
		}
	}

	for {
		select {
		case err := <-w.Errors:
			reload.Log.Println("error watching: ", err)
		case e := <-w.Events:
			switch {
			case e.Has(fsnotify.Create):
				// Watch any created file/directory
				if err := w.Add(e.Name); err != nil {
					log.Printf("error watching %s: %s\n", e.Name, err)
				}
				db(callback(path.Base(e.Name)))

			case e.Has(fsnotify.Write):
				db(callback(path.Base(e.Name)))

			case e.Has(fsnotify.Rename), e.Has(fsnotify.Remove):
				// a renamed file might be outside the specified paths
				directories, _ := recursiveWalk(e.Name)
				for _, v := range directories {
					err := w.Remove(v)
					if err != nil {
						return
					}
				}
				err := w.Remove(e.Name)
				if err != nil {
					return
				}
			}
		}
	}
}

func recursiveWalk(path string) ([]string, error) {
	var res []string
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			res = append(res, path)
		}
		return nil
	})

	return res, err
}
