<script>
import at from '../assets/activitytypes.json'

export default {
  data: function () {
    return {
      activityTypes: at,
      queryString: "",
      fromDate: new Date().toISOString().slice(0, 10),
      toDate: new Date().toISOString().slice(0, 10),
      selectedType: "",
      onlyCached: false,
    };
  }
}
</script>

<template>
  <div id="settings">
    <p>
    <label for="text">Search</label>
    <input id="text" name="text" type="text" v-model="queryString" />
    </p>
    <p>
    <label for="from-date">Start date</label>
    <input id="from-date" name="from-date" type="date" v-model="fromDate"/>
    </p>
    <p>
    <label for="to-date">End date</label>
    <input id="to-date" name="to-date" type="date" v-model="toDate"/>
    </p>
    <p>
    <label for="activity-type">Activity type</label>
    <select id="activity-type" v-model="selectedType">
      <option value="">All</option>
      <option v-for="(name, value) in activityTypes" :value="value">{{ name }}</option>
    </select>
    </p>
    <p>
    <label for="only-cached">Only cached</label>
    <input type="checkbox" id="only-cached" v-model="onlyCached" name="true">
    </p>
    <p>
    <input type="button" value="Get Activities" v-on:click="$emit('submit', $data)"/>
    </p>
    <p>
    <input type="button" value="Clear Shown Activities" v-on:click="$emit('clear')"/>
    </p>
    <p>
    <input type="button" value="Clear Cached Activities" v-on:click="$emit('clearCache')"/>
    </p>
  </div>
</template>
