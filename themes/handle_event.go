/**
 * Copyright (c) 2011 ~ 2013 Deepin, Inc.
 *               2011 ~ 2013 jouyouyun
 *
 * Author:      jouyouyun <jouyouwen717@gmail.com>
 * Maintainer:  jouyouyun <jouyouwen717@gmail.com>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, see <http://www.gnu.org/licenses/>.
 **/

package themes

import (
	"github.com/howeyc/fsnotify"
	"path"
	"regexp"
)

func (obj *Manager) startWatch() {
	if obj.watcher == nil {
		var err error
		obj.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			Logger.Errorf("New Watcher Failed: %v", err)
			panic(err)
		}
	}

	homeDir, _ := objUtil.GetHomeDir()
	obj.watcher.Watch(THEME_SYS_PATH)
	obj.watcher.Watch(path.Join(homeDir, THEME_LOCAL_PATH))
	obj.watcher.Watch(ICON_SYS_PATH)
	obj.watcher.Watch(path.Join(homeDir, ICON_LOCAL_PATH))
	obj.watcher.Watch(SOUND_THEME_PATH)
	obj.watcher.Watch(PERSON_SYS_THEME_PATH)
	obj.watcher.Watch(path.Join(homeDir, PERSON_LOCAL_THEME_PATH))
}

func (obj *Manager) endWatch() {
	if obj.watcher == nil {
		return
	}

	homeDir, _ := objUtil.GetHomeDir()
	obj.watcher.RemoveWatch(THEME_SYS_PATH)
	obj.watcher.RemoveWatch(path.Join(homeDir, THEME_LOCAL_PATH))
	obj.watcher.RemoveWatch(ICON_SYS_PATH)
	obj.watcher.RemoveWatch(path.Join(homeDir, ICON_LOCAL_PATH))
	obj.watcher.RemoveWatch(SOUND_THEME_PATH)
	obj.watcher.RemoveWatch(PERSON_SYS_THEME_PATH)
	obj.watcher.RemoveWatch(path.Join(homeDir, PERSON_LOCAL_THEME_PATH))
}

func (obj *Manager) handleEvent() {
	for {
		select {
		case <-obj.quitFlag:
			return
		case ev, ok := <-obj.watcher.Event:
			if !ok {
				obj.endWatch()
				obj.startWatch()
				break
			}

			if ev == nil {
				break
			}

			ok1, _ := regexp.MatchString(THEME_SYS_PATH, ev.Name)
			ok2, _ := regexp.MatchString(THEME_LOCAL_PATH, ev.Name)
			if ok1 || ok2 {
				obj.setPropGtkThemeList(getGtkThemeList())
			}

			ok1, _ = regexp.MatchString(ICON_SYS_PATH, ev.Name)
			ok2, _ = regexp.MatchString(ICON_LOCAL_PATH, ev.Name)
			if ok1 || ok2 {
				obj.setPropIconThemeList(getIconThemeList())
				obj.setPropCursorThemeList(getCursorThemeList())
			}

			ok1, _ = regexp.MatchString(SOUND_THEME_PATH, ev.Name)
			if ok1 {
				obj.setPropSoundThemeList(getSoundThemeList())
			}

			ok1, _ = regexp.MatchString(PERSON_SYS_THEME_PATH, ev.Name)
			ok2, _ = regexp.MatchString(PERSON_LOCAL_THEME_PATH, ev.Name)
			if ok1 || ok2 {
				obj.setPropThemeList(getDThemeList())
				obj.rebuildThemes()
			}
		case err, ok := <-obj.watcher.Error:
			if !ok || err != nil {
				obj.endWatch()
				obj.startWatch()
			}
		}
	}
}

func (obj *Theme) startWatch() {
	if obj.Type == THEME_TYPE_SYSTEM {
		return
	}

	if obj.watcher == nil {
		var err error
		obj.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			Logger.Errorf("New Watcher Failed: %v", err)
			panic(err)
		}
	}

	filename := path.Join(obj.filePath, "theme.ini")
	obj.watcher.Watch(filename)
}

func (obj *Theme) endWatch() {
	if obj.watcher == nil {
		return
	}

	filename := path.Join(obj.filePath, "theme.ini")
	obj.watcher.RemoveWatch(filename)
}

func (obj *Theme) handleEvent() {
	for {
		select {
		case <-obj.quitFlag:
			return
		case ev, ok := <-obj.watcher.Event:
			if !ok {
				obj.endWatch()
				obj.startWatch()
				break
			}

			if ok, _ := regexp.MatchString(`\.swa?px?$`, ev.Name); ok {
				break
			}

			if ev.IsModify() {
				obj.setAllProps()
			}
		case err, ok := <-obj.watcher.Error:
			if !ok || err != nil {
				obj.endWatch()
				obj.startWatch()
			}
		}
	}
}
