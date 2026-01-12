export interface Comment {
  id: string
  channelId: string
  displayName: string
  message: string
  publishedAt: string
}

export interface CheckedComments {
  [commentId: string]: boolean
}
