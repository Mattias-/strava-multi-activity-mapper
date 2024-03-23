import type { AppState } from "./types.ts";

type ActivityListProps = {
  state: AppState;
};

export default function ActivityList(props: ActivityListProps) {
  const activities = props.state.activities.value.map((a) => (
    <tr>
      <td>{a.type}</td>
      <td>{a.start_date.toISOString().slice(0, 10)}</td>
      <td>
        {a.feature && (
          <a href={a.feature.properties.activity.id}>
            {a.feature.properties.activity.name}
          </a>
        )}
      </td>
    </tr>
  ));
  if (activities === undefined || activities.length === 0) {
    return <></>;
  }
  return (
    <div class="flex shadow-md rounded px-8 pt-2 pb-2 mb-4 max-w-xs bg-gray-100">
      <table>
        <tbody>{activities}</tbody>
      </table>
    </div>
  );
}
