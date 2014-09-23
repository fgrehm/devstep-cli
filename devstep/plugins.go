package devstep

import (
	"github.com/robertkrimen/otto"
	"os"
	"path/filepath"
)

type PluginRuntime interface {
	Trigger(eventName string) error
	Load(pluginPath string) error
}

type pluginRuntime struct {
	projectCfg *ProjectConfig
	vm         *otto.Otto
}

func NewPluginRuntime(projectCfg *ProjectConfig) PluginRuntime {
	runtime := &pluginRuntime{
		projectCfg: projectCfg,
		vm:         otto.New(),
	}

	data := map[string]interface{}{
		"addVolume": runtime.addVolume,
		"addLink":   runtime.addLink,
		"setEnv":    runtime.setEnv,
	}
	obj, err := runtime.vm.ToValue(data)
	if err != nil {
		panic("Error generating config wrapper JS value\n " + err.Error())
	}

	err = runtime.vm.Set("_configWrapper", obj)
	if err != nil {
		panic("Error registering _configWrapper\n " + err.Error())
	}

	_, err = runtime.vm.Run(initPluginJsEnvironment)
	if err != nil {
		panic("Error initializing plugin environment:\n " + err.Error())
	}

	return runtime
}

func (r *pluginRuntime) Trigger(eventName string) error {
	log.Info("Triggering '" + eventName + "' plugin event")
	_, err := r.vm.Run("devstep.trigger('" + eventName + "')")
	return err
}

func (r *pluginRuntime) Load(pluginPath string) error {
	log.Info("Loading '" + pluginPath + "' plugin")

	src, err := os.Open(pluginPath)
	if err != nil {
		panic("Error loading plugin '" + pluginPath + "'\n " + err.Error())
	}

	err = r.vm.Set("_currentPluginPath", filepath.Dir(pluginPath))
	if err != nil {
		panic("Error setting current plugin path _DEVSTEP_ADD_VOLUME\n " + err.Error())
	}

	_, err = r.vm.Run(src)
	return err
}

var initPluginJsEnvironment = `
devstep = {};
devstep._events = { configLoaded: [] };
devstep.trigger = function(eventName) {
	var events = devstep._events[eventName];
	for (var i = 0; i < events.length; i++) {
		events[i](_configWrapper);
	}
};
devstep.on = function(eventName, cb) {
	devstep._events[eventName].push(cb);
};
`

func (r *pluginRuntime) addVolume(call otto.FunctionCall) otto.Value {
	defaults := r.projectCfg.Defaults
	defaults.Volumes = append(defaults.Volumes, call.Argument(0).String())
	return call.This
}

func (r *pluginRuntime) addLink(call otto.FunctionCall) otto.Value {
	defaults := r.projectCfg.Defaults
	defaults.Links = append(defaults.Links, call.Argument(0).String())
	return call.This
}

func (r *pluginRuntime) setEnv(call otto.FunctionCall) otto.Value {
	defaults := r.projectCfg.Defaults
	defaults.Env[call.Argument(0).String()] = call.Argument(1).String()
	return call.This
}
