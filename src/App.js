import Athlete from "./components/Athlete.js";
import QueryForm from "./components/QueryForm.js";
import ActivityList from "./components/ActivityList.js";
import Lmap from "./components/Lmap.js";

import * as L from 'leaflet';

import Dexie from 'dexie';
export const db = new Dexie('myDatabase');
db.version(1).stores({
  activities: '&id, *text, type, start_date',
});

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
    clearActivities: function () {
      this.activities.length = 0;
      var map = this.$refs.map.mapObject;
      map.eachLayer(function (layer) {
        if (Object.prototype.hasOwnProperty.call(layer, "feature")) {
          map.removeLayer(layer);
        }
      });
    },
    getActivities: function (data) {
      var $c = this;

      var text = encodeURIComponent(data.queryString);
      var p = fetch(
        `./activities?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`
      ).then((stream) => stream.json());

      p.then(function (data) {
        data.features.forEach(function (feature){
          var a = feature.properties.activity;
        });
        var a = layer.feature.properties.activity;
        db.activities.put({
          id: String(a.id),
          type: a.type,
          start_date: new Date(a.start_date),
          text: a.name.split(' '),
          layer: layer,
        });
        $c.map.addActivityData(data);
        L.geoJSON(data).eachLayer($c.addToActivityList);
      });

      /*
      db.activities.where('text').anyOfIgnoreCase(data.queryString.split(' ')).distinct().each(function (data) {
        console.log("DB Found:", data)
        $c.$refs.map.addActivityData(data.layer);
        L.geoJSON(data.layer).eachLayer($c.addToActivityList);
      });
      */

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
            v-on:clear="clearActivities"
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
