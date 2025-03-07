export type FilesSelectType = {
	label: string;
	icon: string;
	path: string;
	expandable: string;
	selectable: boolean;
	children: any;
	isDir: boolean;
	lazy?: boolean;
};

export type FilesCodeType = {
	code: String;
	lang: String;
	name: String;
};
