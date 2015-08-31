package appearance

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"pkg.deepin.io/dde/daemon/appearance/background"
	"pkg.deepin.io/dde/daemon/appearance/fonts"
	"pkg.deepin.io/dde/daemon/appearance/subthemes"
	"pkg.deepin.io/lib/graphic"
	dutils "pkg.deepin.io/lib/utils"
	"strings"
)

const (
	// TODO: chdir to 'deepin/dde-daemon/appearance/themes'
	dthemeDir       = "personalization/themes"
	dthemeConfig    = "theme.ini"
	customConfig    = ".local/share/" + dthemeDir + "/Custom/" + dthemeConfig
	defaultFontSize = 10

	kfGroupTheme     = "Theme"
	kfKeyId          = "Id"
	kfKeyName        = "Name"
	kfGroupComponent = "Component"
	kfKeyGtk         = "GtkTheme"
	kfKeyIcon        = "IconTheme"
	kfKeyCursor      = "CursorTheme"
	// TODO: rename to 'Background'
	kfKeyBackground = "BackgroundFile"
	// TODO: rename to 'StandardFont'
	kfKeyStandardFont = "FontName"
	// TODO: rename to 'MonospaceFont'
	kfKeyMonospaceFont = "FontMono"
	kfKeyFontSize      = "FontSize"
)

type dthemeComponent struct {
	Gtk           string
	Icon          string
	Cursor        string
	Background    string
	StandardFont  string
	MonospaceFont string
}

type DTheme struct {
	Id        string
	Name      string
	Path      string
	Thumbnail string

	Previews []string

	Deletable bool

	Gtk    *subthemes.Theme
	Icon   *subthemes.Theme
	Cursor *subthemes.Theme

	Background *background.Background

	StandardFont  *fonts.Family
	MonospaceFont *fonts.Family
	FontSize      int32
}
type DThemes []*DTheme

func ListDTheme() DThemes {
	var infos DThemes
	for _, dir := range getThemeList() {
		info, err := newDThemeFromFile(path.Join(dir, dthemeConfig))
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	return infos
}

func SetDTheme(id string) error {
	dt := ListDTheme().Get(id)
	if dt == nil {
		return fmt.Errorf("Not found '%s'", id)
	}

	subthemes.SetGtkTheme(dt.Gtk.Id)
	subthemes.SetIconTheme(dt.Icon.Id)
	go subthemes.SetCursorTheme(dt.Cursor.Id)
	background.ListBackground().Set(dt.Background.URI)
	fonts.SetFamily(dt.StandardFont.Id, dt.MonospaceFont.Id)
	fonts.SetSize(dt.FontSize)
	return nil
}

func GetDThemeThumbnail(id string) (string, error) {
	dt := ListDTheme().Get(id)
	if dt == nil {
		return "", fmt.Errorf("Invalid dtheme id '%v'", id)
	}
	return dt.Thumbnail, nil
}

func (infos DThemes) GetIds() []string {
	var ids []string
	for _, info := range infos {
		ids = append(ids, info.Id)
	}
	return ids
}

func (infos DThemes) Get(id string) *DTheme {
	for _, info := range infos {
		if info.Id == id {
			return info
		}
	}
	return nil
}

func (infos DThemes) Delete(id string) error {
	info := infos.Get(id)
	if info == nil {
		return fmt.Errorf("Not found '%s'", id)
	}
	return info.Delete()
}

func (infos DThemes) findDThemeId(component *dthemeComponent) string {
	for _, info := range infos {
		if info.isComponentSame(component) {
			return info.Id
		}
	}
	return ""
}

func (info *DTheme) Delete() error {
	if !info.Deletable {
		return fmt.Errorf("Permission Denied")
	}
	return os.RemoveAll(info.Path)
}

func (info *DTheme) isComponentSame(component *dthemeComponent) bool {
	if info.Gtk.Id != component.Gtk ||
		info.Icon.Id != component.Icon ||
		info.Cursor.Id != component.Cursor ||
		info.Background.URI != component.Background ||
		info.StandardFont.Id != component.StandardFont ||
		info.MonospaceFont.Id != component.MonospaceFont {
		return false
	}
	return true
}

func newDThemeFromFile(file string) (*DTheme, error) {
	kfile, err := dutils.NewKeyFileFromFile(file)
	if err != nil {
		return nil, err
	}
	defer kfile.Free()

	var dt DTheme
	id, err := kfile.GetString(kfGroupTheme, kfKeyId)
	if err != nil {
		return nil, err
	}
	if len(id) == 0 {
		return nil, fmt.Errorf("Invalid id")
	}
	dt.Id = id

	name, err := kfile.GetLocaleString(kfGroupTheme, kfKeyId, "\x00")
	if err != nil {
		return nil, err
	}
	dt.Name = name

	tmp, _ := kfile.GetString(kfGroupComponent, kfKeyGtk)
	dt.Gtk = subthemes.ListGtkTheme().Get(tmp)
	if dt.Gtk == nil {
		return nil, fmt.Errorf("Not found gtk theme: %v", tmp)
	}

	tmp, _ = kfile.GetString(kfGroupComponent, kfKeyIcon)
	dt.Icon = subthemes.ListIconTheme().Get(tmp)
	if dt.Icon == nil {
		return nil, fmt.Errorf("Not found icon theme: %v", tmp)
	}

	tmp, _ = kfile.GetString(kfGroupComponent, kfKeyCursor)
	dt.Cursor = subthemes.ListCursorTheme().Get(tmp)
	if dt.Cursor == nil {
		return nil, fmt.Errorf("Not found cursor theme: %v", tmp)
	}

	tmp, _ = kfile.GetString(kfGroupComponent, kfKeyBackground)
	dt.Background = background.ListBackground().Get(tmp)
	if dt.Background == nil {
		return nil, fmt.Errorf("Not found background: %v", tmp)
	}

	tmp, _ = kfile.GetString(kfGroupComponent, kfKeyStandardFont)
	dt.StandardFont = fonts.ListAllFamily().Get(tmp)
	if dt.StandardFont == nil {
		return nil, fmt.Errorf("Not found standard font: %v", tmp)
	}

	tmp, _ = kfile.GetString(kfGroupComponent, kfKeyMonospaceFont)
	dt.MonospaceFont = fonts.ListAllFamily().Get(tmp)
	if dt.MonospaceFont == nil {
		return nil, fmt.Errorf("Not found monospace font: %v", tmp)
	}

	size, err := kfile.GetInteger(kfGroupComponent, kfKeyFontSize)
	if err != nil {
		return nil, err
	}
	dt.FontSize = size

	dt.Path = path.Dir(file)
	dt.Thumbnail = path.Join(dt.Path, "thumbnail.png")
	if !dutils.IsFileExist(dt.Thumbnail) {
		dt.Thumbnail, _ = dt.Background.Thumbnail()
	}
	dt.Previews = getDThemePreviews(dt.Path)
	dt.Deletable = isDeletable(file)

	return &dt, nil
}

func isDeletable(file string) bool {
	if strings.Contains(file, os.Getenv("HOME")) {
		return true
	}
	return false
}

func getDThemePreviews(dir string) []string {
	fr, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer fr.Close()

	names, err := fr.Readdirnames(0)
	if err != nil {
		return nil
	}

	var picts []string
	for _, name := range names {
		if !graphic.IsSupportedImage(path.Join(dir, name)) {
			continue
		}
		picts = append(picts, path.Join(dir, name))
	}
	return picts
}

func getThemeList() []string {
	var list []string
	list = append(list, scanner(path.Join("/usr/share",
		dthemeDir))...)
	list = append(list, scanner(path.Join(os.Getenv("HOME"),
		".local/share", dthemeDir))...)
	return list
}

func scanner(dir string) []string {
	fr, err := os.Open(dir)
	if err != nil {
		return []string{}
	}
	defer fr.Close()

	names, err := fr.Readdirnames(0)
	if err != nil {
		return []string{}
	}

	var ret []string
	for _, name := range names {
		tmp := path.Join(dir, name)
		if !dutils.IsDir(tmp) {
			continue
		}

		if !dutils.IsFileExist(path.Join(tmp, dthemeConfig)) {
			continue
		}

		ret = append(ret, tmp)
	}
	return ret
}

func doWriteCustomDTheme(component *dthemeComponent) error {
	file := path.Join(os.Getenv("HOME"), customConfig)
	err := os.MkdirAll(path.Dir(file), 0755)
	if err != nil {
		return err
	}

	var content string = fmt.Sprintf(`[%s]
%s=Custom
%s=Custom
%s[en_US]=Custom
%s[zh_CN]=自定义
%s[zh_TW]=自定義
%s[zh_HK]=自定義

[%s]
%s=%s
%s=%s
%s=%s
%s=%s
%s=%s
%s=%s
%s=%v`, kfGroupTheme, kfKeyId,
		kfKeyName, kfKeyName, kfKeyName, kfKeyName, kfKeyName,
		kfGroupComponent, kfKeyGtk, component.Gtk,
		kfKeyIcon, component.Icon, kfKeyCursor, component.Cursor,
		kfKeyBackground, component.Background,
		kfKeyStandardFont, component.StandardFont,
		kfKeyMonospaceFont, component.MonospaceFont,
		kfKeyFontSize, defaultFontSize)
	return ioutil.WriteFile(file, []byte(content), 0644)
}