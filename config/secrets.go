/*
 * Copyright 2019 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"bufio"
	"fmt"
	"github.com/99designs/keyring"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)

func ReadUserInput(name string, mask bool) (string, error) {
	fmt.Print("Enter ", name, ": ")
	var val string
	var err error
	if mask {
		raw, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read user input: %v", err)
		}
		val = string(raw)
		fmt.Println()
	} else {
		reader := bufio.NewReader(os.Stdin)
		val, err = reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read user input: %v", err)
		}
		val = strings.Trim(val, "\n")
	}
	return val, nil
}

type Secret interface {
	EnsureKey(name string) (string, error)
	GetKey(name string) (string, error)
	SetKey(name string, value string) error
}

type secret struct {
	ring keyring.Keyring
}

//NewSecretStore instantiates new connection to wallet
func NewSecretStore() (Secret, error) {
	s := &secret{}
	var err error
	s.ring, err = keyring.Open(keyring.Config{
		AllowedBackends:                []keyring.BackendType{keyring.KWalletBackend, keyring.WinCredBackend},
		ServiceName:                    "",
		KeychainName:                   AppName,
		KeychainTrustApplication:       false,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: false,
		KeychainPasswordFunc:           nil,
		FilePasswordFunc:               nil,
		FileDir:                        "",
		KWalletAppID:                   AppName,
		KWalletFolder:                  AppName,
		LibSecretCollectionName:        "",
		PassDir:                        "",
		PassCmd:                        "",
		PassPrefix:                     "",
		WinCredPrefix:                  "",
	})

	if err != nil {
		return s, fmt.Errorf("failed to open wallet: %v", err)
	}
	return s, nil
}

//EnsureKey ensures key exists by either getting it from wallet or asking user for value
func (s *secret) EnsureKey(name string) (string, error) {
	key, err := s.ring.Get(name)
	if err != nil {
		if err.Error() != "unexpected end of JSON input" &&
			err.Error() != "The specified item could not be found in the keyring" {
			panic(fmt.Errorf("failed to read wallet: %v", err))
		}
	}
	var val string
	if len(key.Data) == 0 {
		val, err = ReadUserInput(name, false)
		if err != nil {
			return val, err
		}

		item := keyring.Item{
			Key:  name,
			Data: []byte(val),
		}
		err = s.ring.Set(item)
		if err != nil {
			return "", fmt.Errorf("failed to set wallet key: %v", err)
		}

	} else {
		val = string(key.Data)
	}
	return val, nil
}

//GetKey gets key value from wallet
func (s *secret) GetKey(name string) (string, error) {
	key, err := s.ring.Get(name)
	if err != nil {
		if err.Error() == "unexpected end of JSON input" {
			return "", nil
		} else {
			return "", fmt.Errorf("failed to get wallet value: %v", err)
		}
	}
	value := string(key.Data)
	return value, nil
}

//SetKey stores key value in wallet
func (s *secret) SetKey(name string, value string) error {
	item := keyring.Item{
		Key:  name,
		Data: []byte(value),
	}

	err := s.ring.Set(item)
	if err != nil {
		return fmt.Errorf("failed to set wallet key value: %v", err)
	}
	return nil
}
