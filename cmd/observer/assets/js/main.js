const randomColor = () => {
	let letters = "0123456789ABCDEF".split("")
	let color = "#"
	for (let i = 0; i < 6; ++i) {
		color += letters[Math.floor(Math.random() * 16)]
	}
	return color
}
const randomColors = count => {
	let colors = []
	for (let i = 0; i < count; ++i) {
		colors.push(randomColor())
	}
	return colors
}
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
				backgroundColor: randomColors(Object.keys(timings).length),
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
				backgroundColor: randomColors(vals.length),
			}],
		},
		options: {
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
		},
	})
}
