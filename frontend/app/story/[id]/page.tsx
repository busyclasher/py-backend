import StoryWorkspace from "@/components/StoryWorkspace";

type StoryPageProps = {
  params: { id: string };
};

export default function StoryPage({ params }: StoryPageProps) {
  return <StoryWorkspace id={params.id} />;
}
