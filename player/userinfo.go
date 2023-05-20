package player

import "fmt"

type Userinfo struct {
	Name string
	Skin string
	Hand int
	Rate int
}

func (ui Userinfo) Marshal() string {
	return fmt.Sprintf("\\name\\%s\\skin\\%s\\hand\\%d\\rate\\%d", ui.Name, ui.Skin, ui.Hand, ui.Rate)
}
