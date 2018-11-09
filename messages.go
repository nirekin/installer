package installer

const (

	// Error messages
	ERROR_REQUIRED_ENV           string = "the environment variable \"%s\" should be defined"
	ERROR_PARSING_DESCRIPTOR     string = "Error parsing the descriptor %s, run the \"cli check\" command to get the details"
	ERROR_PARSING_ENVIRONMENT    string = "Error parsing the environment: %s"
	ERROR_LOADING_CLI_PARAMETERS string = "Error loading the CLI parameters : %s"
	ERROR_GENERATING_SSH_KEYS    string = "Error generating the SSH keys %s"
	ERROR_UNSUPORTED_ACTION      string = "the action \"%s\" is not supported by the installer"
	ERROR_CREATING_REPORT_FILE   string = "Error creating the report file  %s"

	ERROR_GENERIC                  string = "An error occurred  %s:"
	ERROR_ADDING_EXCHANGE_FOLDER   string = "An error occurred adding the exchange folder %s: %s"
	ERROR_CREATING_EXCHANGE_FOLDER string = "An error occurred creating the exchange folder %s: %s"
	ERROR_READING_REPORT           string = "error reading the report file \"%s\", error \"%s\""
	ERROR_UNMARSHALLING_REPORT     string = "error Unmarshalling the report file \"%s\", error \"%s\""

	// Log messages
	LOG_STARTING                  string = "Starting the installer..."
	LOG_RUNNING                   string = "Running the installer..."
	LOG_INSTALLER_MODE            string = "Installer in creation mode: %v"
	LOG_CREATION_FOR_CLIENT       string = "Installer creating for the client: %s"
	LOG_PROCESSING_NODE           string = "Processing node: %s"
	LOG_EXTRAVARS_FOR_CLIENT      string = "ExtraVars for client %s : %s"
	LOG_PROCESSING_STACK_PLAYBOOK string = "Processing playbook for stack: %s on node: %s"
	LOG_PROCESSING_STACK_COMPOSE  string = "Processing Docker Compose for stack: %s on node: %s"

	LOG_ACTION_CREATE string = "Action Create asked"
	LOG_ACTION_CHECK  string = "Action Check asked"
	LOG_NO_ACTION     string = "No action specified"

	LOG_SSH_PUBLIC_KEY  string = "Installer using SSH public key: %s"
	LOG_SSH_PRIVATE_KEY string = "Installer using SSH private key: %s"

	LOG_CREATING_UID_FOR_CLIENT string = "Creating a UID %s for client: %s and nodes %s"
	LOG_REUSING_UID_FOR_CLIENT  string = "Reusing the UID %s for client: %s and nodes %s"

	LOG_VALIDATION_LOG_WRITTEN string = "The validation logs have been written into %s\n"
	LOG_VALIDATION_SUCCESSFUL  string = "The envinronment descriptor validation is successful!"

	LOG_REPORT_WRITTEN string = "The execution report file has been written in %s\n"

	LOG_CLI_PARAMS string = "Using CLI parameters: %v"

	LOG_PLATFORM_REPOSITORY   string = "Platform repository %s"
	LOG_PLATFORM_VERSION      string = "Platform version %s"
	LOG_PLATFORM_COMPONENT_ID string = "Platform component ID %s"

	LOG_BUILDING_EXCHANGE_FOLDER string = "Building exchange folder structure"
	LOG_PROVIDER                 string = "Provider %s"

	LOG_RUNNING_SETUP_FOR string = "Running setup for provider %s"
)
