package installer

const (

	// Error messages
	ERROR_REQUIRED_ENV             string = "the environment variable \"%s\" should be defined"
	ERROR_PARSING_DESCRIPTOR       string = "Error parsing the descriptor %s"
	ERROR_PARSING_ENVIRONMENT      string = "Error parsing the environment: %s"
	ERROR_CREATING_EXCHANGE_FOLDER string = "Error creating the exchange folder %s"
	ERROR_GENERATING_SSH_KEYS      string = "Error generating the SSH keys %s"
	ERROR_UNSUPORTED_ACTION        string = "the action \"%s\" is not supported by the installer"
	ERROR_CREATING_REPORT_FILE     string = "Error creating the report file  %s"
	ERROR_GENERIC                  string = "An error occurred  %s:"

	// Log messages
	LOG_STARTING             string = "Starting the installer..."
	LOG_INSTALLER_MODE       string = "Installer in creation mode: %v"
	LOG_CREATION_FOR_CLIENT  string = "Installer creating for the client: %s"
	LOG_PROCESSING_NODE      string = "Processing node: %s"
	LOG_EXTRAVARS_FOR_CLIENT string = "ExtraVars for client %s : %s"

	LOG_SSH_PUBLIC_KEY  string = "Installer using SSH public key: %s"
	LOG_SSH_PRIVATE_KEY string = "Installer using SSH private key: %s"

	LOG_CREATING_UID_FOR_CLIENT string = "Creating a UID %s for client: %s and nodes %s"
	LOG_REUSING_UID_FOR_CLIENT  string = "Reusing the UID %s for client: %s and nodes %s"

	LOG_VALIDATION_LOG_WRITTEN string = "The validation logs have been written into %s\n"
	LOG_VALIDATION_SUCCESSFUL  string = "The envinronment descriptor validation is successful!"
)
