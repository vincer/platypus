package lib

import "github.com/vincer/libhdplatinum"

const MaxShadeHeight = 255

type Response struct {
	Code    int `json:"code"`
	Message string `json:"message"`
}

type ShadeView struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	RoomId string `json:"roomId"`
	Height int `json:"height"`
}

func ShadeViewFromShade(s libhdplatinum.Shade) ShadeView {
	return ShadeView{Id: s.Id(), Name: s.Name(), RoomId: s.RoomId(), Height: s.Height() * 100 / MaxShadeHeight}
}
