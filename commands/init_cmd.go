package commands

import (
	"fmt"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"os"
)

var InitCmd = cli.Command{
	Name: "init",
	Action: func(c *cli.Context) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		configPath := projectRoot + "/devstep.yml"

		if _, err := os.Stat(configPath); err == nil {
			fmt.Println("`devstep.yml` already exists in this directory. Remove it before running `devstep init`.")
			os.Exit(1)
		}
		configFile := []byte(sampleConfig)
		err = ioutil.WriteFile(configPath, configFile, 0755)
		if err != nil {
			fmt.Printf("Error creating config file '%s'\n%s\n", configPath, err)
			os.Exit(1)
		}
		fmt.Printf("Generated sample configuration file in '%s'\n", configPath)
	},
}

var sampleConfig = `# The Docker repository to keep images built by devstep
# DEFAULT: 'devstep/<CURRENT_DIR_NAME>'
# repository: 'repo/name'

# The image used by devstep when building environments from scratch
# DEFAULT: 'fgrehm/devstep:v1.0.0'
# source_image: 'custom/image:tag'

# The host cache dir that gets mounted inside the container at '/home/devstep/cache'
# for speeding up the dependencies installation process.
# DEFAULT: '/tmp/devstep/cache'
# cache_dir: '{{env "HOME"}}/devstep/cache'

# The directory where project sources should be mounted inside the container.
# DEFAULT: '/workspace'
# working_dir: '/home/devstep/gocode/src/github.com/fgrehm/devstep-cli'

# Link to other existing containers (like a database for example).
# Please note that devstep won't start the associated containers automatically
# and an error will be raised in case the linked container does not exist or
# if it is not running.
# DEFAULT: <empty>
# links:
# - "postgres:db"
# - "memcached:mc"

# Additional Docker volumes to share with the container.
# DEFAULT: <empty>
# volumes:
# - "/path/on/host:/path/on/guest"

# Environment variables.
# DEFAULT: <empty>
# environment:
#   RAILS_ENV: "development"

# Custom provisioning steps that can be used when the available buildpacks are not
# enough. Use it to configure addons or run additional commands during the build.
# DEFAULT: <empty>
# provision:
#   - ['configure-addons', 'redis']
#   - ['configure-addons', 'heroku-toolbelt']`
