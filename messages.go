package installer

const (

	// Error messages
	ERROR_REQUIRED_ENV           string = "the environment variable \"%s\" should be defined"
	ERROR_LOADING_CLI_PARAMETERS string = "Error loading the CLI parameters : %s"
	ERROR_GENERATING_SSH_KEYS    string = "Error generating the SSH keys %s"
	ERROR_UNSUPORTED_ACTION      string = "the action \"%s\" is not supported by the installer"

	// Log messages
	LOG_STARTING string = "Starting the installer..."
	LOG_RUNNING  string = "Running the installer..."

	LOG_ACTION_CREATE  string = "Action Create asked"
	LOG_ACTION_INSTALL string = "Action Install asked"
	LOG_ACTION_DEPLOY  string = "Action Deploy asked"
	LOG_ACTION_CHECK   string = "Action Check asked"
	LOG_ACTION_DUMP    string = "Action Dump asked"
	LOG_NO_ACTION      string = "No action specified"

	LOG_SSH_PUBLIC_KEY  string = "Installer using SSH public key: %s"
	LOG_SSH_PRIVATE_KEY string = "Installer using SSH private key: %s"

	LOG_CLI_PARAMS string = "Using CLI parameters: %v"
)
