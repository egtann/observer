const bgColors = [
	"rgba(255, 99, 132, 0.8)",
	"rgba(54, 162, 235, 0.8)",
	"rgba(255, 206, 86, 0.8)",
	"rgba(75, 192, 192, 0.8)",
	"rgba(153, 102, 255, 0.8)",
	"rgba(255, 159, 64, 0.8)",
	"rgba(255, 99, 132, 0.8)",
	"rgba(54, 162, 235, 0.8)",
	"rgba(255, 206, 86, 0.8)",
	"rgba(75, 192, 192, 0.8)",
	"rgba(153, 102, 255, 0.8)",
	"rgba(255, 159, 64, 0.8)",
	"rgba(255, 99, 132, 0.8)",
	"rgba(54, 162, 235, 0.8)",
	"rgba(255, 206, 86, 0.8)",
	"rgba(75, 192, 192, 0.8)",
	"rgba(153, 102, 255, 0.8)",
	"rgba(255, 159, 64, 0.8)",
]
const borderColors = [
	"rgba(255, 99, 132, 1)",
	"rgba(54, 162, 235, 1)",
	"rgba(255, 206, 86, 1)",
	"rgba(75, 192, 192, 1)",
	"rgba(153, 102, 255, 1)",
	"rgba(255, 159, 64, 1)",
	"rgba(255, 99, 132, 1)",
	"rgba(54, 162, 235, 1)",
	"rgba(255, 206, 86, 1)",
	"rgba(75, 192, 192, 1)",
	"rgba(153, 102, 255, 1)",
	"rgba(255, 159, 64, 1)",
	"rgba(255, 99, 132, 1)",
	"rgba(54, 162, 235, 1)",
	"rgba(255, 206, 86, 1)",
	"rgba(75, 192, 192, 1)",
	"rgba(153, 102, 255, 1)",
	"rgba(255, 159, 64, 1)",
]
const barChartOpts = {
	legend: {
		display: false,
	},
	scales: {
		yAxes: [{
			ticks: {
				beginAtZero: true,
			},
		}],
	},
}
function loadChart(el, timings) {
	new Chart(el, {
		type: "bar",
		data: {
			labels: Object.keys(timings),
			datasets: [{
				label: "ms",
				data: Object.values(timings),
				backgroundColor: bgColors,
				borderColor: borderColors,
			}],
		},
		options: barChartOpts,
	})
}
function loadMsgChart(el, timings) {
	let keys = []
	let vals = []
	for (let i = 0; i < timings.length; ++i) {
		let timing = timings[i]
		let key = timing.Msg
		if (key.length > 22) {
			key = key.slice(0, 22) + "..."
		}
		keys.push(key)

		// Convert from nano to ms
		timing.Duration /= 1000000
		vals.push(timing.Duration)
	}
	new Chart(el, {
		type: "bar",
		data: {
			labels: keys,
			datasets: [{
				label: "ms",
				data: vals,
				backgroundColor: bgColors,
				borderColor: borderColors,
			}],
		},
		options: {
			legend: {
				display: false,
			},
			scales: {
				yAxes: [{
					type: "logarithmic",
					ticks: {
						beginAtZero: true,
					},
				}],
			},
		},
	})
}
