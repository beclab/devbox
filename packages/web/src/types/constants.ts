export const authLevelOptions = ['private', 'public'];

export const visiblityOptions = ['True', 'False'];

export const mfaLevelOptions = ['two_factor', 'one_factor', 'public', 'deny'];

export const requiredCpuOptions = ['core', 'm'];

export const requiredOptions = ['Ei', 'Pi', 'Ti', 'Gi', 'Mi', 'Ki'];

export const middlewareOptions = ['postgres', 'redis', 'mongodb'];

export const databasesOptions = ['Enable', 'Disable'];

export const dependencyOptions = ['system', 'application'];

export const validDuration = ['ns', 'us', 'ms', 's', 'm', 'h'];

export const statusStyle = {
	pending: {
		color: '#ADADAD'
	},
	processing: {
		color: '#FFC46D'
	},
	installing: {
		color: '#3377FF'
	},
	upgrading: {
		color: '#3377FF'
	},
	running: {
		color: '#29CC5F'
	},
	uninstalling: {
		color: '#FF4D4D'
	},
	errored: {
		color: '#FF4D4D'
	},
	failed: {
		color: '#FF4D4D'
	},
	canceled: {
		color: '#ADADAD'
	},
	suspend: {
		color: '#e53935'
	}
};

export enum OPERATE_ACTION {
	ADD_FOLDER = 'ADD_FOLDER',
	ADD_FILE = 'ADD_FILE',
	RENAME = 'RENAME',
	DELETE = 'DELETE'
}

export enum CONFIG_TAB {
	METADATA = 'METADATA',
	DETAILS = 'DETAILS',
	SPACE = 'SPACE',
	PERMISSIONS = 'PERMISSIONS',
	OPTIONS = 'OPTIONS'
}

// export const FilesOption: Record<
//   OPERATE_ACTION,
//   {
//     name: string;
//     icon: string;
//   }
// > = {
//   [OPERATE_ACTION.ADD_FOLDER]: {
//     name: 'Add Folder',
//     icon: 'sym_r_create_new_folder',
//   },
//   [OPERATE_ACTION.ADD_FILE]: {
//     name: 'Add File',
//     icon: 'sym_r_note_add',
//   },
//   [OPERATE_ACTION.RENAME]: {
//     name: 'Rename',
//     icon: 'sym_r_edit_square',
//   },
//   [OPERATE_ACTION.DELETE]: {
//     name: 'Delete',
//     icon: 'sym_r_delete',
//   },
// };
