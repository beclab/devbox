// This is just an example,
// so you can safely delete all default props below

export default {
	launch_input_placehoder: 'Search',
	launch_no_result: 'No Results',
	page_404: 'Oops. Nothing here...',
	no_data: 'No data.',

	home_welcome: 'Welcome to DevBox',
	home_desc: 'An easy way to develop in Terminus',
	home_start: 'Start',
	home_create: 'Create a new application',
	home_update: 'Upload a chart package',
	home_recent: 'Recent',
	home_recent_resc_1: 'The application you recently',
	home_recent_resc_2: 'developed will be displayed here.',

	home_documents: 'Documents',
	home_doc_1: 'DevBox Tutorial',
	home_doc_2: 'Introduction to Basic Concepts of Terminus Application',
	home_doc_3: 'Learn about Terminus Application Chart',
	home_doc_4: 'Configuration Guideline for TerminusManifest',
	home_doc_5: 'DevBox TutorialTerminus Market protocol overview',
	home_doc_6: 'How to sbmit an application to the Terminus Market',
	home_visit_1: 'Visit',
	home_visit_2: 'for more information.',
	home_appname: 'App Name',
	home_appname_hint: 'App’s namespace in Terminus system.',
	home_appname_rules_1: 'Please input the app name.',
	home_appname_rules_2: 'must start with an alphabetic character.',
	home_appname_rules_3: 'must contain only lowercase alphanumeric characters.',
	home_apptype: 'App Type',
	home_apptype_hint: 'Choose application type.',
	home_entrance_port: 'Main Entrance Port',
	home_entrance_port_hint: 'Port of main entrance.',
	home_entrance_port_rules_1: 'Port of main entrance.',
	home_entrance_port_rules_2: 'must be an int from 0 to 65535',
	home_image: 'Image',
	home_image_hint: 'Image for app containers.',
	home_image_rules: 'Please input the image',
	home_port: 'Image',
	home_port_hint: 'Specify ports that need to be exposed.',
	home_port_rules: 'must be an int from 0 to 65535',
	home_memory: 'Required Memory',
	home_memory_hint: 'Requested memory resources for the app.',
	home_memory_rules: 'must be a number greater than 0.',

	home_gpu: 'Required GPU',
	home_gpu_hint: 'Requested GPU memory resources for the app.',
	home_gpu_place: 'Leave empty if no GPU required.',
	cancel: 'Cancel',
	create: 'Create',
	submit: 'Submit',
	app_files: 'Files',
	app_containers: 'Containers',
	app_config: 'Config',
	type: 'Type',
	name: 'Name',
	version: 'Version',
	username: 'Username',
	password: 'Password',
	distributed: 'Distributed',
	operations: 'Operations',
	required: 'Required',
	limits: 'Limits',
	upload: 'Upload',
	replace: 'Replace',
	close: 'Close',
	policies: 'Policies',

	btn_install: 'Install',
	btn_installing: 'Installing',
	btn_upgrade: 'Upgrade',
	btn_cancel: 'Cancel',
	btn_confirm: 'Confirm',
	btn_preview: 'Preview',
	btn_upload: 'Upload',
	btn_download: 'Download',
	btn_uninstall: 'Uninstall',
	btn_delete: 'Delete',
	btn_unbind: 'Unbind',
	btn_bind: 'Bind',
	btn_binding: 'Binding',
	btn_open_ide: 'Open IDE',
	btn_rename: 'Rename',

	application: 'Application',
	container_list: 'Dev Container List',

	requirementNotSpecified: 'requirement not specified',
	upload_file_nofi: 'Drag and drop JPEG, PNG or WEBP files here to upload',
	upload_icon_nofi: 'Drag and drop a PNG or WEBP files here to upload',

	dialog_create_file: 'Create File',
	dialog_create_folder: 'Create Folder',
	dialog_create_title: 'Name',
	dialog_title_bind: 'Bind a dev container',
	dialog: {
		addEntrance: 'Add Entrance',
		editEntrance: 'Edit Entrance',
		addSubPolicies: 'Add Sub Policies',
		editSubPolicies: 'Edit Sub Policies',
		addClientReference: 'Add Client Reference',
		editClientReference: 'Edit Client Reference',
		addRequiredData: 'Add Required Data',
		editRequiredData: 'Edit Required Data',
		addClient: 'Add {type} Client',
		editClient: 'Edit {type} Client',
		addMiddle: 'Add {type}',
		editMiddle: 'Edit {type}'
	},

	containers_env: 'Env',
	containers_bind_app: 'Binding App Container',
	containers_bind_dev: 'Binding Dev Container',
	containers_pod_selector: 'Pod Selector',
	containers_update_time: 'Update Time',
	containers_dev_env: 'Dev Env',
	containers_select_env: 'Select dev containers',
	containers_input_name: 'Container Name',

	config_name: 'App Configuration',
	config_metadata_icon: 'Icon',
	config_metadata_icon_hint:
		"Your app icon appears in the Terminus Market. The app's icon must be in PNG or WEBP format, up to 512 KB, with a size of 256x256 px.",
	config_metadata_apptitle: 'App Title',
	config_metadata_apptitle_hint: 'Your app title appears in the app market.',
	config_metadata_apptitle_rules: 'Please input the app title',
	config_metadata_versionname: 'Version Name',
	config_metadata_versionname_hint:
		"Your app's version displayed in the Terminus Market. Please specify in the SemVer 2.0.0 format.",
	config_metadata_versionname_rules: 'Please input the version name',
	config_metadata_categories: 'Categories',
	config_metadata_categories_hint:
		'Used to display your app on different category pages in the Terminus Market.',
	config_metadata_categories_rules: 'Please input categories',
	config_metadata_shortdesc: 'Short Description',
	config_metadata_shortdesc_hint:
		'A short description appears below app title in the Terminus Market.',
	config_metadata_shortdesc_rules: 'Please input the short description',
	config_metadata_fulldesc: 'Full Description',
	config_metadata_fulldesc_hint: 'A full description of your app.',
	config_metadata_fulldesc_rules: 'Please input the full description',

	config_details_upgrade_desc: 'Upgrade Description',
	config_details_upgrade_hint: 'Describe what is new in this upgraded version.',
	config_details_upgrade_rules: 'Please input the upgrade description',
	config_details_developer: 'Developer',
	config_details_developer_hint: 'The name of developer of this app.',
	config_details_developer_rules: 'Please input the developer',
	config_details_submitter: 'Submitter',
	config_details_submitter_hint:
		'The name of submitter who submits this app to the app market.',
	config_details_submitter_rules: 'Please input the submitter',
	config_details_featimage: 'Featured Image',
	config_details_featimage_hint:
		'Upload a featured image for the app. The image must be in JPEG, PNG or WEBP format, up to 8MB each, with a size of 1440x900 px.',
	config_details_promotemage: 'Promote Image',
	config_details_promotemage_hint:
		'Upload 2-8 app screenshots for promotion. Screenshots must be in JPEG, PNG or WEBP format, up to 8MB each, with a size of 1440x900 px.',
	config_details_document: 'Document',
	config_details_document_hint:
		'Add a link to the documents or user manual for your app.',
	config_details_website: 'Website',
	config_details_website_hint:
		'Add a link to your official website, if you have one.',
	config_details_legalnote: 'Legal Note',
	config_details_legalnote_hint:
		'Add a link to the legal notes that you want to display on the app market.',
	config_details_license: 'License',
	config_details_license_hint: "Add a link to your app's license agreement.",
	config_details_sourcecode: 'Source Code',
	config_details_sourcecode_hint: "Add a link to your app's source code.",
	config_details_supportclient: 'Support Client',
	config_details_supportclient_desc:
		'Add links to your app clients on other platforms.',

	config_space_entrances: 'Entrances',
	config_space_entrances_desc:
		'Specify how to access this app, at least 1 required.',
	config_space_resources: 'Resources',
	config_space_resources_desc:
		'Specify requested and limited resources for your app.',
	config_space_middleware: 'Middleware',
	config_space_middleware_desc: 'Add the necessary middleware for your app.',
	config_space_dependencies: 'Dependencies',
	config_space_dependencies_desc:
		'Indicate if your app depends on other apps or requires a specific OS version.',
	config_space_appdata: 'Require App Data',
	config_space_adddata_desc:
		'Requires read and write permissions to appdata directory.',
	config_space_systemdata: 'Require System Data',
	config_space_systemdata_desc:
		'Require permissions to access system data through service providers.',

	config_space_entrancename: 'Entrance Name',
	config_space_entrancename_hint: 'Assign a unique name for this entrance.',
	config_space_entrancename_rules: 'Please input the entrance name',
	config_space_entrancename_rules2:
		'must contain only lowercase alphanumeric characters and hyphens.',

	config_space_addentrance_name_hint: 'The app name of dependent app.',
	config_space_addentrance_name_rules1: 'Please input the app name.',
	config_space_addentrance_name_rules2:
		'must start with an alphabetic character.',
	config_space_addentrance_name_rules3:
		'must contain only lowercase alphanumeric characters.',
	config_space_addentrance_version_hint: 'Required version.',
	config_space_addentrance_version_rules: 'Please input the version',

	config_space_entrancetitle: 'Entrance Title',
	config_space_entrancetitle_hint:
		'Title that appears on the Terminus desktop after installation.',
	config_space_entrancetitle_rules: 'Please input the entrance title',

	config_space_entranceicon: 'Entrance Icon',
	config_space_entranceicon_desc:
		'Icon that appears in the Terminus desktop after installed.',

	config_space_hostname: 'Host Name',
	config_space_hostname_hint: 'Ingress name for this entrance.',
	config_space_hostname_rules: 'Please input the host name',
	config_space_hostname_rule2:
		'must contain only lowercase alphanumeric characters and hyphens.',

	config_space_gpu_required: 'GPU Required',
	config_space_required: 'Required {type}',
	config_space_limited: 'Limited {type}',
	config_space_required_hint: 'Minimum {type} required for the app.',
	config_space_limited_hint:
		'{type} limit for the app. The app will be suspended if the resource limit is exceeded.',

	config_space_postgres_name_rules: 'Please input the required username',
	config_space_postgres_password_place:
		'Leave empty to generate a 16-bit random password',

	config_space_middleware_Databases: 'Databases',

	config_option_cluster: 'Cluster Scoped',
	config_option_cluster_desc:
		'Whether this app is installed for all users in a Terminus cluster.',
	config_option_Reference_name_rules: 'Please input the name',
	config_option_Reference_name_rules2:
		'must start with an alphabetic character.',
	config_option_Reference_name_rules3:
		'must contain only lowercase alphanumeric characters.',

	config_space_client: 'Client Reference',
	config_space_client_desc:
		'Specify the client apps that need to access this cluster app.',
	config_space_analytics: 'Enable Analytics',
	config_space_analytics_desc: 'Enable website analytics for your app.',
	config_space_websocket: 'Enable Websocket',
	config_space_websocket_desc: 'Enable websocket for your app.',
	config_space_port: 'Port',
	config_space_url: 'URL',
	config_space_visible: 'Visible',
	config_space_visible_hint:
		'Show entrance icon and title on the Terminus desktop.',
	config_space_authlevel: 'Auth Level',
	config_space_authlevel_desc:
		'A private entrance requires activating Tailscale for access.',
	config_space_openmethod: 'Open Method',
	config_space_openmethod_desc:
		'Show entrance icon and title on the Terminus desktop.',

	config_space_policy: 'Policy Scope',
	config_space_policy_hint:
		'Set the affected domain of this policy.  Regular expressions are supported.',
	config_space_policy_place:
		'Add effected URLs of the policy, regular expression supported',
	config_space_policy_rules: 'Please input the effected domain',

	config_space_mfalevel: 'MFA Level',
	config_space_mfalevel_hint:
		'Two-Factor requires additional credentials with an OTP (One-TimePassword) to access the entrance.',

	config_space_onetimevalid: 'One Time Valid',
	config_space_onetimevalid_hint:
		'Authentication is required every time to access this entrance.',

	config_space_validduration: 'Valid Duration',
	config_space_validduration_hint:
		'Set the time period (in seconds) before a user is asked to MFA again. Leave empty for one time valid.',

	config_space_description: 'Description',
	config_space_description_hint: 'A brief description of this policy.',

	config_permissions_datagroup: 'Data Group',
	config_permissions_datagroup_hint: 'Group of required data.',
	config_permissions_datagroup_rules: 'Please input the data group',

	config_permissions_datatype: 'Data type',
	config_permissions_datatype_hint: 'Type of required data.',
	config_permissions_datatype_rules: 'Please input the data type',

	config_permissions_version_hint: 'Version of required data.',
	config_permissions_version_rules: 'Please input the version',
	config_permissions_operations_hint:
		'Specify required service provider operations.',
	config_permissions_operations_rules: 'Please enter key input the operations.',

	config: {
		addClients: 'Add Clients',
		addEntrance: 'Add Entrance',
		add: 'Add',
		addDependencies: 'Add Dependencies'
	},

	enums: {
		ADD_FOLDER: 'Add Folder',
		ADD_FILE: 'Add File',
		RENAME: 'Rename',
		DELETE: 'Delete',
		METADATA: 'Metadata',
		DETAILS: 'Details',
		SPACE: 'Specs',
		PERMISSIONS: 'Permissions',
		OPTIONS: 'Options',
		DevBox: 'DevBox',
		Home: 'Home',
		Containers: 'Containers',
		Help: 'Help',
		Applications: 'Applications'
	}
};
