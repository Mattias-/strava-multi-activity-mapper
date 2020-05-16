export default {
    props: ["activities"],
    template: `
        <div id="activity-list">
          <table>
            <tbody>
              <tr v-for="a in sortedActivities">
                <td><a :href="a.url">{{ a.name }}</a></td>
                <td>{{ a.start_date_local | isoDay }}</td>
                <td>{{ a.type }}</td>
              </tr>
            </tbody>
          </table>
        </div>
`,
    filters: {
        isoDay: function(d) {
            return new Date(d).toISOString().slice(0, 10);
        }
    },
    computed: {
        sortedActivities() {
            return this.activities.slice(0).sort((a, b) => new Date(a.start_date) < new Date(b.start_date) ? 1 : -1)
        }
    }
}
