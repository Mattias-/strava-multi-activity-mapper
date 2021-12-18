import Athlete from "./components/Athlete.js";
import QueryForm from "./components/QueryForm.js";
import ActivityList from "./components/ActivityList.js";
import Lmap from "./components/Lmap.js";

import * as L from 'leaflet';
import 'leaflet/dist/leaflet.css';

export default {
  data() {
    return {
      activityTypes: {},
      athlete: {},
      activities: [],
    };
  },
  components: {
    Lmap,
    Athlete,
    QueryForm,
    ActivityList,
  },
  computed: {
    authed() {
      return this.athlete != null;
    },
  },
  methods: {
    getActivities: function (data) {
      var $c = this;
      var text = encodeURIComponent(data.queryString);
      var p = fetch(
        `./activities?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`
      ).then((stream) => stream.json());
      p.then(this.$refs.map.addActivityData);
      p.then(function (data) {
        L.geoJSON(data).eachLayer($c.addToActivityList);
      });
    },

    addToActivityList: function (layer) {
      if (Object.prototype.hasOwnProperty.call(layer, "feature")) {
        var a = layer.feature.properties.activity;
        a.url = "https://strava.com/activities/" + a.id;
        if (!this.activities.find((b) => b.id == a.id)) {
          this.activities.push(a);
        }
      }
    },
  },
  mounted() {
    fetch(`./athlete`)
      .then((stream) => stream.json())
      .then((data) => (this.athlete = data))
      .catch(() => (this.athlete = null));
    fetch("/activitytypes.json")
      .then((stream) => stream.json())
      .then((data) => (this.activityTypes = data));
  },
  template: `
      <div id="menu">
        <template v-if="authed">
          <athlete v-bind:athlete="athlete"></athlete>
          <query-form
            v-bind:activity-types="activityTypes"
            v-on:submit="getActivities"
          ></query-form>
          <activity-list v-bind:activities="activities"></activity-list>
        </template>
        <template v-else>
          <a href="auth" class="menu-image"
            ><img src="/connectwith_strava.svg" alt="Connect with Strava"
          /></a>
        </template>
        <a href="https://www.strava.com" class="menu-image"
          ><img src="/poweredby_strava.svg" alt="Powered by Strava"
        /></a>
      </div>
      <Lmap ref="map"></Lmap>
`,
};
