"use client";

import { useEffect, useMemo, useState } from "react";
import useSWR from "swr";
import {
  appendBlock,
  executeStory,
  getStory,
  leaveComment,
  openStoryEventStream,
  Story,
} from "@/lib/api";

interface StoryWorkspaceProps {
  id: string;
}

export default function StoryWorkspace({ id }: StoryWorkspaceProps) {
  const { data, error, isLoading, mutate } = useSWR(["story", id], () => getStory(id));
  const story = data;
  const [newBlock, setNewBlock] = useState({
    type: "markdown" as const,
    source: "### Next Steps\nDetail the experiment plan.",
  });
  const [commentDraft, setCommentDraft] = useState({
    author: "analyst@example.com",
    body: "Love this insight!",
  });
  const [isExecuting, setIsExecuting] = useState(false);
  const lastUpdated = useMemo(() => {
    if (!story) return "";
    return new Date(story.updatedAt).toLocaleString();
  }, [story]);

  useEffect(() => {
    if (!story) {
      return;
    }
    const source = openStoryEventStream(id, () => {
      mutate();
    });
    return () => {
      source.close();
    };
  }, [id, story, mutate]);

  const handleAddBlock = async () => {
    if (!story) return;
    await appendBlock(story.id, {
      type: newBlock.type,
      source: newBlock.source,
    });
    setNewBlock((prev) => ({ ...prev, source: "" }));
    mutate();
  };

  const handleComment = async () => {
    if (!story || commentDraft.body.trim() === "") return;
    await leaveComment(story.id, commentDraft);
    setCommentDraft((draft) => ({ ...draft, body: "" }));
    mutate();
  };

  const handleExecute = async () => {
    if (!story) return;
    setIsExecuting(true);
    try {
      await executeStory(story.id, "team.bot@example.com");
      mutate();
    } finally {
      setIsExecuting(false);
    }
  };

  if (isLoading) {
    return <p className="text-sm text-slate-500">Loading story…</p>;
  }

  if (error) {
    return <p className="rounded-md bg-red-100 p-4 text-sm text-red-700">Failed to load story.</p>;
  }

  if (!story) {
    return <p className="text-sm text-slate-500">Story not found.</p>;
  }

  return (
    <div className="flex flex-col gap-6">
      <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <header className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h2 className="text-xl font-semibold text-slate-900">{story.title}</h2>
            <p className="text-sm text-slate-600">{story.description}</p>
          </div>
          <div className="flex gap-2 text-xs text-slate-500">
            <span>Revision {story.revisionId}</span>
            <span>Updated {lastUpdated}</span>
          </div>
        </header>
      </section>

      <section className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <article className="flex flex-col gap-4">
          {story.blocks.map((block) => (
            <BlockView key={block.id} block={block} />
          ))}
          <div className="rounded-lg border border-dashed border-slate-300 bg-white p-4">
            <h3 className="mb-2 text-sm font-semibold text-slate-700">Add block</h3>
            <div className="flex flex-col gap-3">
              <select
                value={newBlock.type}
                onChange={(event) => setNewBlock((prev) => ({ ...prev, type: event.target.value as typeof prev.type }))}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              >
                <option value="markdown">Markdown</option>
                <option value="code">Code</option>
                <option value="visualization">Visualization</option>
              </select>
              <textarea
                value={newBlock.source}
                onChange={(event) => setNewBlock((prev) => ({ ...prev, source: event.target.value }))}
                rows={4}
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none"
              />
              <button
                onClick={handleAddBlock}
                className="w-fit rounded-md bg-primary-600 px-4 py-2 text-sm font-medium text-white shadow hover:bg-primary-700"
              >
                Append block
              </button>
            </div>
          </div>
        </article>

        <aside className="flex flex-col gap-4">
          <div className="rounded-lg border border-slate-200 bg-white p-4">
            <h3 className="mb-3 text-sm font-semibold text-slate-700">Execution</h3>
            <button
              onClick={handleExecute}
              disabled={isExecuting}
              className="w-full rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700 disabled:opacity-60"
            >
              {isExecuting ? "Running…" : "Run notebook"}
            </button>
          </div>
          <div className="rounded-lg border border-slate-200 bg-white p-4">
            <h3 className="mb-3 text-sm font-semibold text-slate-700">Comments</h3>
            <div className="flex flex-col gap-3">
              {story.comments.map((comment) => (
                <div key={comment.id} className="rounded border border-slate-100 bg-slate-50 p-3 text-sm">
                  <div className="flex justify-between text-xs text-slate-500">
                    <span>{comment.author}</span>
                    <span>{new Date(comment.createdAt).toLocaleString()}</span>
                  </div>
                  <p className="mt-2 text-slate-700">{comment.body}</p>
                </div>
              ))}
              <textarea
                value={commentDraft.body}
                onChange={(event) => setCommentDraft((draft) => ({ ...draft, body: event.target.value }))}
                rows={3}
                placeholder="Share feedback for the team"
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none"
              />
              <button
                onClick={handleComment}
                className="w-fit rounded-md border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 hover:border-slate-300"
              >
                Post comment
              </button>
            </div>
          </div>
        </aside>
      </section>
    </div>
  );
}

function BlockView({ block }: { block: Story["blocks"][number] }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-center justify-between text-xs uppercase tracking-wide text-slate-500">
        <span>{block.type}</span>
        <span>#{block.position + 1}</span>
      </div>
      <pre className="mt-3 whitespace-pre-wrap text-sm text-slate-800">{block.source}</pre>
      {block.outputs.length > 0 && (
        <div className="mt-4 rounded border border-slate-100 bg-slate-50 p-3 text-xs text-slate-600">
          <p className="mb-1 font-medium text-slate-700">Outputs</p>
          {block.outputs.map((output, index) => (
            <pre key={index} className="whitespace-pre-wrap text-xs text-slate-700">
              {output.data}
            </pre>
          ))}
        </div>
      )}
    </div>
  );
}
