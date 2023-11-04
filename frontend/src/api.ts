import { Athlete, ApiActivities } from "./types.ts";

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;

export async function getAthlete(): Promise<Athlete> {
  const response = await fetch(`${apiBaseUrl}/athlete`, {
    credentials: "include",
  });
  if (response.status === 200) {
    return await response.json();
  } else {
    throw new Error(
      `API request failed with status: ${response.status} - ${response.statusText}`,
    );
  }
}

export interface GetActivitiesProps {
  [index: string]: string;
  queryString: string;
  fromDate: string;
  toDate: string;
  type: string;
}

export async function getActivities(
  data: GetActivitiesProps,
): Promise<ApiActivities> {
  const queryString = new URLSearchParams({
    q: data.queryString,
    after: data.fromDate,
    before: data.toDate,
    type: data.type,
  }).toString();
  const response = await fetch(`${apiBaseUrl}/activities?${queryString}`, {
    credentials: "include",
  });
  if (response.status === 200) {
    return await response.json();
  } else {
    throw new Error(
      `API request failed with status: ${response.status} - ${response.statusText}`,
    );
  }
}

export type ATs = {
  [key: string]: string;
};

export async function getActivityTypes(): Promise<ATs> {
  const response = await fetch(`${apiBaseUrl}/activitytypes.json`);
  if (response.status === 200) {
    return await response.json();
  } else {
    throw new Error(
      `API request failed with status: ${response.status} - ${response.statusText}`,
    );
  }
}
