"use client";

import { useState } from "react";
import useSWR, { mutate } from "swr";
import { createStory, listStories, Story } from "@/lib/api";

const fetcher = () => listStories();

export default function StoriesOverview() {
  const { data, error, isLoading } = useSWR("stories", fetcher, {
    refreshInterval: 15000,
  });
  const [isCreating, setIsCreating] = useState(false);
  const [formState, setFormState] = useState({
    title: "Q2 Revenue Review",
    description: "Outline the story behind the latest forecast.",
    owners: "data.team@example.com",
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const stories = data ?? [];

  const handleCreate = async () => {
    setIsSubmitting(true);
    try {
      const owners = formState.owners
        .split(",")
        .map((owner) => owner.trim())
        .filter(Boolean);
      await createStory({
        title: formState.title,
        description: formState.description,
        owners,
        visibility: "organization",
        tags: ["finance", "quarterly"],
        blocks: [
          {
            type: "markdown",
            source: "# Executive Summary\nSummarise key takeaways for stakeholders.",
            position: 0,
          },
          {
            type: "code",
            language: "python",
            source: "# TODO: Pull numbers from the warehouse\n",
            position: 1,
          },
        ],
      });
      setIsCreating(false);
      mutate("stories");
    } catch (err) {
      console.error(err);
      alert("Unable to create story. Check console for details.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-semibold text-slate-900">Workspace</h2>
          <p className="text-sm text-slate-600">Recent narratives and active experiments.</p>
        </div>
        <button
          onClick={() => setIsCreating(true)}
          className="inline-flex items-center justify-center rounded-md bg-primary-600 px-4 py-2 text-sm font-medium text-white shadow hover:bg-primary-700 focus:outline-none"
        >
          New Story
        </button>
      </div>

      {error && <p className="rounded-md bg-red-100 p-4 text-sm text-red-700">Failed to load stories.</p>}
      {isLoading && <p className="text-sm text-slate-500">Loading stories…</p>}

      <div className="grid gap-4 md:grid-cols-2">
        {stories.map((story) => (
          <StoryCard key={story.id} story={story} />
        ))}
        {!isLoading && stories.length === 0 && (
          <div className="rounded-lg border border-dashed border-slate-300 p-8 text-center text-sm text-slate-500">
            No stories yet. Create one to invite collaborators.
          </div>
        )}
      </div>

      {isCreating && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/30 p-4">
          <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-slate-900">Create story</h3>
            <div className="flex flex-col gap-4">
              <label className="text-sm font-medium text-slate-700">
                Title
                <input
                  value={formState.title}
                  onChange={(event) => setFormState((state) => ({ ...state, title: event.target.value }))}
                  className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none"
                />
              </label>
              <label className="text-sm font-medium text-slate-700">
                Description
                <textarea
                  value={formState.description}
                  onChange={(event) => setFormState((state) => ({ ...state, description: event.target.value }))}
                  rows={3}
                  className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none"
                />
              </label>
              <label className="text-sm font-medium text-slate-700">
                Owners (comma separated)
                <input
                  value={formState.owners}
                  onChange={(event) => setFormState((state) => ({ ...state, owners: event.target.value }))}
                  className="mt-2 w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none"
                />
              </label>
            </div>
            <div className="mt-6 flex justify-end gap-3 text-sm">
              <button onClick={() => setIsCreating(false)} className="rounded-md border border-slate-200 px-4 py-2 text-slate-600 hover:border-slate-300">
                Cancel
              </button>
              <button
                disabled={isSubmitting}
                onClick={handleCreate}
                className="rounded-md bg-primary-600 px-4 py-2 font-medium text-white shadow hover:bg-primary-700 disabled:opacity-60"
              >
                {isSubmitting ? "Creating…" : "Create"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function StoryCard({ story }: { story: Story }) {
  return (
    <article className="flex flex-col gap-4 rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h3 className="text-lg font-semibold text-slate-900">{story.title}</h3>
          <p className="mt-1 text-sm text-slate-600">{story.description}</p>
        </div>
        <span className="rounded bg-slate-100 px-2 py-1 text-xs font-medium uppercase tracking-wide text-slate-600">
          {story.visibility}
        </span>
      </div>
      <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
        <span>Owners: {story.owners.join(", ")}</span>
        <span>Blocks: {story.blocks.length}</span>
        <span>Updated {new Date(story.updatedAt).toLocaleString()}</span>
      </div>
      <a
        href={`/story/${story.id}`}
        className="inline-flex w-fit items-center gap-2 text-sm font-medium text-primary-600 hover:underline"
      >
        Open story ?
      </a>
    </article>
  );
}
