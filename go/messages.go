package main

const (

	// Error messages
	ERROR_REQUIRED_ENV             string = "the environment variable \"%s\" should be defined"
	ERROR_PARSING_DESCRIPTOR       string = "Error parsing the descriptor %s"
	ERROR_CREATING_EXCHANGE_FOLDER string = "Error creating the exchange folder %s"

	// Log messages
	LOG_STARTING             string = "Starting the installer..."
	LOG_INSTALLER_MODE       string = "Installer in creation mode: %v"
	LOG_CREATION_FOR_CLIENT  string = "Installer creating for the client: %s"
	LOG_PROCESSING_NODE      string = "Processing node: %s"
	LOG_EXTRAVARS_FOR_CLIENT string = "ExtraVars for client %s : %s"

	LOG_CREATING_UID_FOR_CLIENT string = "Creating a UID %s for client: %s and nodes %s"
	LOG_REUSING_UID_FOR_CLIENT  string = "Reusing the UID %s for client: %s and nodes %s"
)
