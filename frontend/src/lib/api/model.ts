export type Group = {
	id: number;
	name: string;
	auto_fetch_full_content?: boolean;
};

export type Feed = {
	id: number;
	name: string;
	link: string;
	failure: string;
	updated_at: Date;
	suspended: boolean;
	auto_fetch_full_content?: boolean;
	req_proxy: string;
	unread_count: number;
	group: Group;
};

export type Item = {
	id: number;
	title: string;
	link: string;
	content: string;
	full_content?: string;
	unread: boolean;
	bookmark: boolean;
	pub_date: Date;
	updated_at: Date;
	feed: Pick<Feed, 'id' | 'name' | 'link'>;
};
