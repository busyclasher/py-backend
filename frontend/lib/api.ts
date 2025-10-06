export type BlockType = "markdown" | "code" | "visualization";

export interface Output {
  kind: string;
  mimeType: string;
  data: string;
}

export interface Block {
  id: string;
  type: BlockType;
  language?: string;
  source: string;
  position: number;
  outputs: Output[];
}

export interface Comment {
  id: string;
  storyId: string;
  blockId?: string;
  author: string;
  body: string;
  createdAt: string;
}

export interface Story {
  id: string;
  title: string;
  description: string;
  owners: string[];
  visibility: "private" | "organization" | "public";
  revisionId: string;
  blocks: Block[];
  comments: Comment[];
  tags: string[];
  createdAt: string;
  updatedAt: string;
}

export interface ExecutionResult {
  storyId: string;
  revision: string;
  startedAt: string;
  finishedAt: string;
  status: string;
  blocks: Block[];
  logs: string[];
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!response.ok) {
    throw new Error(`API ${response.status} ${response.statusText}`);
  }
  return response.json() as Promise<T>;
}

export function listStories() {
  return request<Story[]>("/api/stories");
}

export function getStory(storyId: string) {
  return request<Story>(`/api/stories/${storyId}`);
}

export function createStory(payload: {
  title: string;
  description: string;
  owners: string[];
  visibility: Story["visibility"];
  tags: string[];
  blocks: Array<{
    type: BlockType;
    language?: string;
    source: string;
    position?: number;
  }>;
}) {
  return request<Story>("/api/stories", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function appendBlock(storyId: string, block: { type: BlockType; language?: string; source: string; position?: number }) {
  return request<Story>(`/api/stories/${storyId}/blocks`, {
    method: "POST",
    body: JSON.stringify(block),
  });
}

export function leaveComment(storyId: string, payload: { author: string; body: string; blockId?: string }) {
  return request<Story>(`/api/stories/${storyId}/comments`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function executeStory(storyId: string, actor: string) {
  return request<ExecutionResult>(`/api/stories/${storyId}/execute`, {
    method: "POST",
    body: JSON.stringify({ actor }),
  });
}

export function openStoryEventStream(storyId: string, onMessage: (event: MessageEvent) => void) {
  const url = `${API_BASE}/api/stories/${storyId}/events`;
  const source = new EventSource(url);
  source.onmessage = onMessage;
  return source;
}
