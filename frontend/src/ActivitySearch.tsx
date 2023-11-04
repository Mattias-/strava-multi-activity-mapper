import { useEffect, useState, useMemo } from "preact/hooks";
import { AppState, Activity, Feature } from "./types.ts";
import { db } from "./db.ts";
import { getActivityTypes, ATs, getActivities } from "./api.ts";

type ActivitySearchProps = {
  state: AppState;
};

export default function ActivitySearch(props: ActivitySearchProps) {
  const data = {
    queryString: "",
    fromDate: "",
    toDate: "",
    type: "",
  };
  var onlyCached = false;

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

  function addActivityToDb(feature: Feature) {
    const a = feature.properties.activity;
    const nameA = a.name?.split(" ");
    const descA = a.description?.split(" ") || [];
    const text = [...nameA, ...descA];
    const ac: Activity = {
      id: a.id,
      type: a.type,
      start_date: new Date(a.start_date),
      text,
      feature,
    };
    db.activities.put(ac);
  }

  async function getActs() {
    if (!onlyCached) {
      const p = await getActivities(data);
      if (p !== undefined) {
        p.features.forEach(addActivityToDb);
      }
    }

    var activities = db.activities.toCollection();
    if (data.queryString) {
      activities = db.activities
        .where("text")
        .anyOfIgnoreCase(data.queryString.split(" "))
        .distinct();
    }

    activities
      .and((value) => {
        var lower = new Date(data.fromDate);
        lower.setHours(0, 0, 0);
        var upper = new Date(data.toDate);
        upper.setHours(23, 59, 59);
        return value.start_date >= lower && value.start_date <= upper;
      })
      .each(function (data: Activity) {
        console.log("DB Found:", data);
        const feature = data.feature;
        const a = feature.properties.activity;
        const ac: Activity = {
          id: a.id,
          type: a.type,
          start_date: new Date(a.start_date),
          text: [],
          feature,
        };
        props.state.activities.value = [...props.state.activities.value, ac];
      });
  }

  return (
    <div>
      <form class="flex flex-col bg-gray shadow-md rounded px-8 pt-4 pb-4 mb-4 max-w-xs bg-gray-100">
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
            type="button"
            onClick={() => getActs()}
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
            Clear Shown
          </button>
        </div>
        <div class="mb-4">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
            type="button"
          >
            Clear Cache
          </button>
        </div>
      </form>
      <a href="https://www.strava.com" class="menu-image">
        <img src="/poweredby_strava.svg" alt="Powered by Strava" />
      </a>
    </div>
  );
}
