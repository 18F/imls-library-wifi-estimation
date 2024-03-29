<template>
	<h1 class="header">
		{{fcfsId}} Wi-Fi statistics on {{day}}
	</h1>
	<div class="body" v-if="isLoading">
		<p class="text">
			Loading...
		</p>
	</div>
	<div class="body" v-else>
		<p class="text">
			{{totalDevices}} devices seen
		</p>
		<p class="text">
			{{totalPatrons}} patrons
		</p>
		<p class="text">
			{{totalMinutes}} minutes served
		</p>
	</div>
	<div class="footer">
		<div class="button" @click="day = previousDay(day)">
			&larr; Previous day
		</div>
		<div class="button" @click="day = nextDay(day)">
			Next day &rarr;
		</div>
	</div>
</template>

<script>
 import { watch, ref, onMounted } from 'vue';
 import { useApi } from '@directus/extensions-sdk';
 import dayjs from 'dayjs';

 // dayjs is lightweight, so add in plugins to support timezone and custom parsing
 import utc from 'dayjs/plugin/utc';
 import timezone from 'dayjs/plugin/timezone';
 import customParseFormat from 'dayjs/plugin/customParseFormat';
 dayjs.extend(utc);
 dayjs.extend(timezone);
 dayjs.extend(customParseFormat);
 dayjs.tz.setDefault('America/Los_Angeles');

 export default {
	 props: {
		 day: {
			 type: Date,
			 default: '2021-08-31', // just happens to be tax day!
		 },
		 fcfsId: {
			 type: String,
			 default: 'CA5678-999',
			 required: true,
		 },
		 tableName: {
			 type: String,
			 default: 'durations',
			 required: true,
		 },
	 },
	 setup(props) {
		 var totalDevices = ref(0);
		 var totalPatrons = ref(0);
		 var totalMinutes = ref(0);
		 var isLoading = ref(true);
		 const api = useApi();

		 watch([
			 () => props.day,
		 ],
			   () => {
				   retrieveMetrics();
		 });

		 onMounted(retrieveMetrics);

		 return {
			 totalDevices,
			 totalPatrons,
			 totalMinutes,
			 isLoading,
			 fcfsId: props.fcfsId,
			 previousDay,
			 nextDay,
		 };

		 async function retrieveMetrics() {
			 isLoading.value = true;
			 const filter = {
				 _and: [
					 {
						 "start": {
						 	 _gte: unixEpoch(props.day),
						 }
					 },
					 {
						 "end": {
						 	 _lt: unixEpoch(nextDay(props.day)),
						 }
					 },
					 {
						 "fcfs_seq_id": {
							 _eq: props.fcfsId,
						 }
					 }
				 ]
			 };
			 // NB. these calculations are NOT ACCURATE and are only for
			 // prototyping purposes.
			 const devices = await api.get(`/items/${props.tableName}`, {
				 params: {
					 aggregate: {
						 "count": "*",
					 },
					 filter,
				 },
			 });
			 totalDevices.value = parseInt(devices.data.data[0].count).toLocaleString();
			 const patrons = await api.get(`/items/${props.tableName}`, {
				 params: {
					 aggregate: {
						 "countDistinct": "patron_index",
					 },
					 filter,
				 },
			 });
			 totalPatrons.value = parseInt(patrons.data.data[0].countDistinct.patron_index).toLocaleString();
			 const times = await api.get(`/items/${props.tableName}`, {
				 params: {
					 aggregate: {
						 "sum": ["end", "start"],
					 },
					 filter,
				 },
			 });
			 // this is a hack. we want a proper "duration" column since there
			 // does not seem to be a way to do custom selects using the API.
			 const seconds = (times.data.data[0].sum.end - times.data.data[0].sum.start);
			 totalMinutes.value = Math.trunc(seconds / 60).toLocaleString();
			 isLoading.value = false;
		 }

		 function unixEpoch(date) {
			 return dayjs(date).unix();
		 }

		 function previousDay(date) {
			 return dayjs(date).subtract(1, 'day').format('YYYY-MM-DD');
		 }

		 function nextDay(date) {
			 return dayjs(date).add(1, 'day').format('YYYY-MM-DD');
		 }
	 },
 };
</script>

<style scoped>
 .header {
	 display: flex;
	 margin: 0 1rem;
	 font-weight: bold;
 }

 .text {
	 padding: 8px 0;
 }

 .body {
	 margin: 1rem;
	 min-height: 9rem; /* prevent "load" flicker */
 }

 .footer {
	 margin: 0 1rem;
	 display: flex;
	 flex-direction: row;
	 justify-content: space-between;
 }

 .button {
	 border: 1px solid #777;
	 border-radius: 5px;
	 cursor: pointer;
	 padding: 0.25rem 0.5rem;
	 background: #efefef;
 }
</style>
