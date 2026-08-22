// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	desktopassets "github.com/Its-Satyajit/reqly/apps/desktop"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	app := application.New(application.Options{
		Name:        "reqly",
		Description: "A local-first, Git-native API development environment",
		Services: []application.Service{
			application.NewService(NewAppService()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(desktopassets.FrontendDist),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Reqly",
		Width:            1280,
		Height:           800,
		MinWidth:         940,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(14, 15, 18),
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
