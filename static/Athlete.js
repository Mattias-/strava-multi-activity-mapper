export default {
  props: ["athlete"],
  template: `
        <div id="athlete">
          <a :href="athlete.url">
          <img :alt="athlete.name" class="avatar-img" :src="athlete.image">
          <h2>{{ athlete.name }}</h2>
          </a>
        </div>
`,
};
