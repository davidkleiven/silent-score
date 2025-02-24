package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("5"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	pad2              = lipgloss.NewStyle().MarginLeft(2)
	errorColor        = lipgloss.Color("1")
	okColor           = lipgloss.Color("2")
	headerStyle       = lipgloss.NewStyle().PaddingLeft(2).Background(lipgloss.Color("4")).Bold(true)
)
