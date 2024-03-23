import Papa from "papaparse";
import { useEffect, useMemo, useState } from "preact/hooks";
import {
  type ATs,
  getActivities,
  getActivity,
  getActivityTypes,
} from "./api.ts";
import { db } from "./db.ts";
import type { Activity, AppState, Feature } from "./types.ts";

type ActivitySearchProps = {
  state: AppState;
};

export default function ActivitySearch(props: ActivitySearchProps) {
  const data = {
    queryString: "",
    fromDate: new Date(Date.now() - 7 * 24 * 3600 * 1000)
      .toISOString()
      .slice(0, 10),
    toDate: new Date().toISOString().slice(0, 10),
    type: "",
  };
  let onlyCached = false;

  const [activityTypes, setActivityTypes] = useState<ATs>({});
  useEffect(() => {
    (async () => {
      const at = await getActivityTypes();
      if (at !== undefined) {
        setActivityTypes(at);
      }
    })();
  }, []);

  const atOpts = useMemo(() => {
    return Object.entries(activityTypes).map(([k, v]) => (
      <option value={k}>{v}</option>
    ));
  }, [activityTypes]);

  function clearActivities() {
    props.state.activities.value = [];
  }

  function featureToActivity(feature: Feature): Activity {
    const a = feature.properties.activity;
    const nameA = a.name?.split(" ");
    const descA = a.description?.split(" ") || [];
    const text = [...nameA, ...descA];
    return {
      id: a.id,
      type: a.type,
      start_date: new Date(a.start_date),
      text,
      feature,
    };
  }

  async function getActs() {
    if (!onlyCached) {
      const p = await getActivities(data);
      const as = p.features.map(featureToActivity);
      for (const ac of as) {
        db.activities.put(ac);
      }
    }

    let activities = db.activities.toCollection();
    if (data.queryString) {
      activities = db.activities
        .where("text")
        .anyOfIgnoreCase(data.queryString.split(" "))
        .distinct();
    }

    const lower = new Date(data.fromDate);
    lower.setHours(0, 0, 0);
    const upper = new Date(data.toDate);
    upper.setHours(23, 59, 59);

    activities
      .and((value: Activity) => {
        return value.start_date >= lower && value.start_date <= upper;
      })
      .modify((data: Activity) => {
        if (data.feature === undefined) {
          console.log("Found in DB with NO feature:", data);
          getActivity(data.id)
            .then((a) => {
              console.log("Got activity from API", a);
              data.feature = a.features[0];
              data.start_date = new Date(
                a.features[0].properties.activity.start_date,
              );
            })
            .catch((e) => {
              console.error("Error when getting single activity", e);
            });
        } else {
          console.log("Found in DB with feature:", data);
        }
        props.state.activities.value = [...props.state.activities.value, data];
      });
  }

  function onSubmit(e: Event) {
    e.preventDefault();
    getActs();
  }

  interface MyWindow {
    showOpenFilePicker(): Promise<FileSystemFileHandle[]>;
  }

  async function uploadCSV() {
    (window as unknown as MyWindow)
      .showOpenFilePicker()
      .then((fhs) => fhs[0].getFile())
      .then((file) =>
        Papa.parse<Record<string, string>>(file, {
          skipEmptyLines: true,
          header: true,
          transformHeader: (header) => {
            // Some headers contain a span element, just get the content of these.
            if (header.startsWith("<span")) {
              return header.split(">")[1].split("<")[0];
            }
            return header;
          },
          step: (res) => {
            const id = res.data["Activity ID"];
            const type = res.data["Activity Type"];
            const start_date = new Date(res.data["Activity Date"]);
            const name = res.data["Activity Name"];
            const description = res.data["Activity Description"];
            const nameA = name?.split(" ");
            const descA = description?.split(" ") || [];
            const text = [...nameA, ...descA];
            const ac: Activity = {
              id: id,
              type: type,
              start_date: start_date,
              text,
            };
            db.activities.put(ac);
          },
        }),
      );
  }

  return (
    <div>
      <form
        onSubmit={onSubmit}
        class="flex flex-col bg-gray shadow-md rounded px-8 pt-4 pb-4 mb-4 max-w-xs bg-gray-100"
      >
        <div class="mb-4">
          <label class="block text-gray-700 text-sm font-bold mb-2" for="text">
            Search
          </label>
          <input
            class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
            id="text"
            type="text"
            placeholder="#k"
            value={data.queryString}
            onChange={(e) => {
              if (e.target instanceof HTMLInputElement) {
                data.queryString = e.target.value;
              }
            }}
          />
        </div>
        <div class="mb-4">
          <label
            for="from-date"
            class="block text-gray-700 text-sm font-bold mb-2"
          >
            Start date
          </label>
          <input
            id="from-date"
            name="from-date"
            type="date"
            class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
            value={data.fromDate}
            onChange={(e) => {
              if (e.target instanceof HTMLInputElement) {
                data.fromDate = e.target.value;
              }
            }}
          />
        </div>
        <div class="mb-4">
          <label
            for="to-date"
            class="block text-gray-700 text-sm font-bold mb-2"
          >
            End date
          </label>
          <input
            id="to-date"
            name="to-date"
            type="date"
            class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
            value={data.toDate}
            onChange={(e) => {
              if (e.target instanceof HTMLInputElement) {
                data.toDate = e.target.value;
              }
            }}
          />
        </div>
        <div class="mb-4">
          <label
            for="activity-type"
            class="block text-gray-700 text-sm font-bold mb-2"
          >
            Activity type
          </label>
          <select
            id="activity-type"
            class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
            value={data.type}
            onChange={(e) => {
              if (e.target instanceof HTMLSelectElement) {
                data.type = e.target.value;
              }
            }}
          >
            <option value="">All</option>
            {atOpts}
          </select>
        </div>
        <div class="mb-4">
          <label
            for="only-cached"
            class="block text-gray-700 text-sm font-bold mb-2"
          >
            Only cached
          </label>
          <input
            type="checkbox"
            id="only-cached"
            name="true"
            checked={onlyCached}
            onClick={() => {
              onlyCached = !onlyCached;
            }}
          />
        </div>
        <div class="mb-4">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
            type="submit"
          >
            Get Activities
          </button>
        </div>
        <div class="mb-4">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
            type="button"
            onClick={() => clearActivities()}
          >
            Clear
          </button>
        </div>
        <div class="mb-4">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
            type="button"
            onClick={() => uploadCSV()}
          >
            Upload exported activities CSV
          </button>
        </div>
      </form>
      <a href="https://www.strava.com" class="menu-image">
        <img src="/poweredby_strava.svg" alt="Powered by Strava" />
      </a>
    </div>
  );
}
