import AthleteCard from "./AthleteCard";
import ActivitySearch from "./ActivitySearch.tsx";
import MapComponent from "./MapComponent";
import ActivityList from "./ActivityList.tsx";
import { signal } from "@preact/signals";
import { AppState, Activity } from "./types.ts";

export function App() {
  const state: AppState = { activities: signal<Activity[]>([]) };
  return (
    <>
      <MapComponent state={state} />
      <div
        class="flex flex-col p-4 max-w-md h-screen overflow-y-scroll absolute left-0 top-0"
        style="z-index:1000;"
      >
        <AthleteCard />
        <ActivitySearch state={state} />
        <ActivityList state={state} />
      </div>
    </>
  );
}
