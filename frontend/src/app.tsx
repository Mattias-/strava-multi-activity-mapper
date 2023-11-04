import AthleteComponent from "./Athlete.tsx";
import ActivitySearch from "./ActivitySearch.tsx";
import Map from "./Map.tsx";
import ActivityList from "./ActivityList.tsx";
import { signal } from "@preact/signals";
import { AppState, Activity } from "./types.ts";

export function App() {
  const state: AppState = { activities: signal<Activity[]>([]) };
  return (
    <>
      <Map state={state} />
      <div
        class="flex flex-col p-4 max-w-md h-screen overflow-y-scroll absolute left-0 top-0"
        style="z-index:1000;"
      >
        <AthleteComponent />
        <ActivitySearch state={state} />
        <ActivityList state={state} />
      </div>
    </>
  );
}
