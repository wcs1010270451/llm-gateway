export interface ListResponse<T> {
  items: T[];
  total: number;
}

export interface PageResponse<T> extends ListResponse<T> {
  page: number;
  page_size: number;
}
