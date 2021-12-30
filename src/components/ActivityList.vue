<script>
export default {
  props: ["activities"],
  computed: {
    sortedActivities() {
      return this.activities
        .slice(0)
        .sort((a, b) =>
          new Date(a.start_date) < new Date(b.start_date) ? 1 : -1
        );
    },
  },
};
</script>

<template>
  <div id="activity-list">
    <table>
      <tbody>
        <tr v-for="a in sortedActivities">
          <td><a :href="a.url">{{ a.name }}</a></td>
          <td>{{ new Date(a.start_date_local).toISOString().slice(0, 10) }}</td>
          <td>{{ a.type }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
