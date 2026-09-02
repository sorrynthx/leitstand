package tui

import (
	"fmt"
	"leitstand/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
)

type AuthType int

const (
	AuthTypePassword AuthType = 0
	AuthTypeKey      AuthType = 1
)

type HostFormData struct {
	Name       string
	Address    string
	Port       int
	Username   string
	AuthMethod string
	Password   string
	KeyPath    string
	KeyContent string
	Passphrase string
	Group      string
}

type HostForm struct {
	isEditMode   bool
	hostID       int64
	existingKey  string
	existingPass string
	authType     AuthType
	inputs       []textinput.Model
	focusIndex   int
	width        int
	height       int
	errMessage   string
	filePicker   *FilePickerModal
}

func NewHostForm() *HostForm {
	inputs := make([]textinput.Model, 8)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "e.g. prod-api-01"
	inputs[0].Focus()
	inputs[0].Prompt = "Host Name:       "
	inputs[0].Width = 38

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "e.g. 192.168.0.50 or host.example.com"
	inputs[1].Prompt = "IP / Address:    "
	inputs[1].Width = 38

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "22"
	inputs[2].Prompt = "Port:            "
	inputs[2].SetValue("22")
	inputs[2].Width = 10

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "e.g. ubuntu, root, devops"
	inputs[3].Prompt = "Username:        "
	inputs[3].Width = 38

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "SSH Password"
	inputs[4].Prompt = "Password:        "
	inputs[4].EchoMode = textinput.EchoPassword
	inputs[4].EchoCharacter = '•'
	inputs[4].Width = 38

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "~/.ssh/id_rsa or click Browse"
	inputs[5].Prompt = "Private Key Path:"
	inputs[5].Width = 38

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Key Passphrase (optional)"
	inputs[6].Prompt = "Passphrase:      "
	inputs[6].EchoMode = textinput.EchoPassword
	inputs[6].EchoCharacter = '•'
	inputs[6].Width = 38

	inputs[7] = textinput.New()
	inputs[7].Placeholder = "e.g. Production Web, Staging"
	inputs[7].Prompt = "Group Name:      "
	inputs[7].Width = 38

	return &HostForm{
		isEditMode: false,
		authType:   AuthTypePassword,
		inputs:     inputs,
		focusIndex: 0,
	}
}

func NewEditHostForm(host *storage.Host, secret *storage.HostSecret, payload *storage.SecretPayload) *HostForm {
	form := NewHostForm()
	form.isEditMode = true
	if host != nil {
		form.hostID = host.ID
		form.inputs[0].SetValue(host.Name)
		form.inputs[1].SetValue(host.Address)
		form.inputs[2].SetValue(fmt.Sprintf("%d", host.Port))
		form.inputs[3].SetValue(host.Username)
		form.inputs[7].SetValue(host.GroupName)
	}
	if secret != nil {
		if secret.AuthMethod == "private_key" {
			form.authType = AuthTypeKey
		} else {
			form.authType = AuthTypePassword
		}
	}
	if payload != nil {
		if payload.Password != "" {
			form.inputs[4].SetValue(payload.Password)
			form.existingPass = payload.Password
		}
		if payload.PrivateKey != "" {
			if payload.KeyPath != "" {
				form.inputs[5].SetValue(payload.KeyPath)
			} else {
				form.inputs[5].SetValue("[Encrypted Vault Key Loaded]")
			}
			form.existingKey = payload.PrivateKey
		}
		if payload.Passphrase != "" {
			form.inputs[6].SetValue(payload.Passphrase)
		}
	}
	return form
}

func (f *HostForm) getFieldOrder() []int {
	if f.authType == AuthTypeKey {
		return []int{0, 1, 2, 3, -1, 5, 6, 7}
	}
	return []int{0, 1, 2, 3, -1, 4, 7}
}
