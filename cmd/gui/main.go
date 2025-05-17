package main

import (
	"fmt"
	"path/filepath"
	"projectson/config"
	"projectson/ui"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const preferenceCurrentConfig = "currentConfigPath_v2"

type AppState struct {
	fyneApp             fyne.App
	mainWindow          fyne.Window
	collectorService    *ui.CollectorService
	statusBar           *widget.Label
	tabs                *container.AppTabs
	configPageMaker     func() fyne.CanvasObject
	exclusionsPageMaker func() fyne.CanvasObject
	configDocsPageMaker func() fyne.CanvasObject // New field for config docs page
	previewPageMaker    func() fyne.CanvasObject
	runPageMaker        func() fyne.CanvasObject
	statsPageMaker      func() fyne.CanvasObject
}

func main() {
	myApp := app.NewWithID("looqey.projectson.go")
	myWindow := myApp.NewWindow("ProjectSon")

	initialCfg := config.NewDefaultConfig()
	collectorSvc := ui.NewCollectorService(initialCfg, myApp, myWindow)

	appState := &AppState{
		fyneApp:          myApp,
		mainWindow:       myWindow,
		collectorService: collectorSvc,
		statusBar:        widget.NewLabel("Ready. Load or configure."),
	}

	appState.configPageMaker = func() fyne.CanvasObject {
		return ui.MakeConfigPage(appState.collectorService, func() {
			appState.statusBar.SetText("Configuration applied. Collector will use updated settings.")
			appState.fullUIUpdateOnConfigChange()
		})
	}
	appState.exclusionsPageMaker = func() fyne.CanvasObject {
		return ui.MakeExclusionsPage(appState.collectorService, func() {
			appState.statusBar.SetText("Content exclusion rules applied. Collector will use updated settings.")
			appState.fullUIUpdateOnConfigChange()
		})
	}
	appState.configDocsPageMaker = func() fyne.CanvasObject { // New page maker initialization
		return ui.MakeConfigDocsPage()
	}
	appState.previewPageMaker = func() fyne.CanvasObject {
		return ui.MakePreviewPage(appState.collectorService, myWindow)
	}
	appState.runPageMaker = func() fyne.CanvasObject {
		return ui.MakeRunPage(appState.collectorService, myWindow, appState.statusBar, func() {
			appState.refreshStatsPagePostRun()
		})
	}
	appState.statsPageMaker = func() fyne.CanvasObject {
		return ui.MakeStatsPage(appState.collectorService)
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.FileIcon(), appState.loadConfigDialog),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), appState.saveConfigDialog),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			dialog.ShowInformation("About", "projectson by looqey", appState.mainWindow)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.LogoutIcon(), func() {
			dialog.ShowConfirm("Exit", "Are you sure you want to close FileCollector Go?", func(confirm bool) {
				if confirm {
					appState.fyneApp.Quit()
				}
			}, appState.mainWindow)
		}),
	)

	appState.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Config", theme.SettingsIcon(), appState.configPageMaker()),
		container.NewTabItemWithIcon("Exclusions", theme.ContentCutIcon(), appState.exclusionsPageMaker()),
		container.NewTabItemWithIcon("Config Docs", theme.HelpIcon(), appState.configDocsPageMaker()), // New tab item
		container.NewTabItemWithIcon("Preview", theme.SearchIcon(), appState.previewPageMaker()),
		container.NewTabItemWithIcon("Run", theme.MediaPlayIcon(), appState.runPageMaker()),
		container.NewTabItemWithIcon("Stats", theme.ListIcon(), appState.statsPageMaker()),
	)
	appState.tabs.SetTabLocation(container.TabLocationLeading)

	appState.loadInitialConfig()

	content := container.NewBorder(toolbar, appState.statusBar, nil, nil, appState.tabs)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.SetMaster()
	myWindow.SetCloseIntercept(func() {
		dialog.ShowConfirm("Exit", "Are you sure you want to exit?", func(confirm bool) {
			if confirm {
				myWindow.Close()
			}
		}, myWindow)
	})
	myWindow.ShowAndRun()
}

func (s *AppState) fullUIUpdateOnConfigChange() {
	// This call ensures that the service acknowledges the config from a central point if needed,
	// but primarily, pages are rebuilt using the latest config from the service.
	// s.collectorService.UpdateConfig(s.collectorService.GetConfig()) // Redundant if called by pages

	tabItemsCount := 6 // Updated count (Config, Exclusions, Config Docs, Preview, Run, Stats)
	if s.tabs != nil && len(s.tabs.Items) == tabItemsCount {
		s.tabs.Items[0].Content = s.configPageMaker()
		s.tabs.Items[1].Content = s.exclusionsPageMaker()
		s.tabs.Items[2].Content = s.configDocsPageMaker() // Update for new tab
		s.tabs.Items[3].Content = s.previewPageMaker()    // Index shifted
		s.tabs.Items[4].Content = s.runPageMaker()        // Index shifted
		s.tabs.Items[5].Content = s.statsPageMaker()      // Index shifted
		s.tabs.Refresh()
		selected := s.tabs.Selected()
		if selected != nil {
			s.tabs.Select(selected) // Attempt to re-select the current tab to maintain context
		}
	} else {
		fmt.Printf("Warning: fullUIUpdateOnConfigChange called but tabs are not fully initialized (expected %d tabs, got %d).\\n", tabItemsCount, len(s.tabs.Items))
	}
	fmt.Println("Full UI refreshed due to config change mechanism.")
}

func (s *AppState) refreshStatsPagePostRun() {
	stats := s.collectorService.GetLastRunStats()
	if stats.ErrorMessage != "" {
		s.statusBar.SetText(fmt.Sprintf("Run failed: %s", stats.ErrorMessage))
	} else {
		s.statusBar.SetText(fmt.Sprintf("Run completed. Output: %s. Files: %d", stats.OutputSize, stats.FileCount))
	}

	tabItemsCount := 6 // Updated count
	// Ensure stats tab exists (it's the last one)
	if s.tabs != nil && len(s.tabs.Items) == tabItemsCount {
		s.tabs.Items[tabItemsCount-1].Content = s.statsPageMaker() // Refresh stats page (still the last one)
		s.tabs.Refresh()
	}
	fmt.Println("Stats page refreshed after run.")
}

func (s *AppState) loadInitialConfig() {
	prefPath := s.fyneApp.Preferences().String(preferenceCurrentConfig)
	initialStatus := "Using default configuration. Create or load a new one."
	if prefPath != "" {
		cfg, err := config.LoadConfig(prefPath)
		if err == nil {
			s.collectorService.UpdateConfig(cfg)
			initialStatus = "Loaded saved config: " + filepath.Base(prefPath)
			s.fullUIUpdateOnConfigChange()
			s.statusBar.SetText(initialStatus)
			return
		}
		s.fyneApp.Preferences().RemoveValue(preferenceCurrentConfig) // Remove invalid pref
		initialStatus = "Failed to load saved config, using defaults."
	}
	s.statusBar.SetText(initialStatus)
}

func (s *AppState) loadConfigDialog() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, s.mainWindow)
			s.statusBar.SetText("Error during file open dialog.")
			return
		}
		if reader == nil {
			s.statusBar.SetText("Load cancelled.")
			return
		}
		defer reader.Close()
		filePath := reader.URI().Path()
		if !strings.HasSuffix(strings.ToLower(filePath), ".yaml") && !strings.HasSuffix(strings.ToLower(filePath), ".yml") {
			dialog.ShowError(fmt.Errorf("selected file is not a YAML file (.yaml or .yml): %s", filepath.Base(filePath)), s.mainWindow)
			s.statusBar.SetText("Invalid file type selected.")
			return
		}
		cfg, loadErr := config.LoadConfig(filePath)
		if loadErr != nil {
			dialog.ShowError(loadErr, s.mainWindow)
			s.statusBar.SetText("Error loading config: " + loadErr.Error())
			return
		}
		s.collectorService.UpdateConfig(cfg)
		s.fyneApp.Preferences().SetString(preferenceCurrentConfig, filePath)
		s.statusBar.SetText("Loaded config: " + filepath.Base(filePath))
		s.fullUIUpdateOnConfigChange()
	}, s.mainWindow)
	fileDialog.Show()
}

func (s *AppState) saveConfigDialog() {
	currentCfgToSave := s.collectorService.GetConfig()
	if err := currentCfgToSave.Validate(); err != nil {
		dialog.ShowError(fmt.Errorf("cannot save, configuration is invalid: %w", err), s.mainWindow)
		s.statusBar.SetText("Config invalid. Please fix before saving.")
		return
	}

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, s.mainWindow)
			s.statusBar.SetText("Error during file save dialog.")
			return
		}
		if writer == nil {
			s.statusBar.SetText("Save cancelled.")
			return
		}
		defer writer.Close()
		filePathToSave := writer.URI().Path()
		if !strings.HasSuffix(strings.ToLower(filePathToSave), ".yaml") && !strings.HasSuffix(strings.ToLower(filePathToSave), ".yml") {
			filePathToSave += ".yaml" // Ensure .yaml extension
		}
		saveErr := currentCfgToSave.SaveConfig(filePathToSave)
		if saveErr != nil {
			dialog.ShowError(saveErr, s.mainWindow)
			s.statusBar.SetText("Error saving config: " + saveErr.Error())
			return
		}
		s.fyneApp.Preferences().SetString(preferenceCurrentConfig, filePathToSave)
		s.statusBar.SetText("Saved config to: " + filepath.Base(filePathToSave))
	}, s.mainWindow)

	suggestedName := "projectson_config.yaml"
	prefPath := s.fyneApp.Preferences().String(preferenceCurrentConfig)
	if prefPath != "" {
		suggestedName = filepath.Base(prefPath)
	}
	fileDialog.SetFileName(suggestedName)
	fileDialog.Show()
}
