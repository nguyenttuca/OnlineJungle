const daysOfWeek = ["Thứ 2", "Thứ 3", "Thứ 4", "Thứ 5", "Thứ 6", "Thứ 7", "Chủ nhật"];
const rawData = `[\u007b\u0022day\u0022: \u0022Thứ 5\u0022, \u0022end\u0022: \u002211:00\u0022, \u0022start\u0022: \u002210:50\u0022, \u0022subject\u0022: \u0022Tú\u0022\u007d]`;
const weeklySchedule = JSON.parse(rawData);

let html = '';
let currentDay = '';
const sorted = [...weeklySchedule].sort((a, b) => {
    const dayA = daysOfWeek.indexOf(a.day);
    const dayB = daysOfWeek.indexOf(b.day);
    if (dayA !== dayB) return dayA - dayB;
    return a.start.localeCompare(b.start);
});

sorted.forEach(slot => {
    if (slot.day !== currentDay) {
        if (currentDay !== '') html += `</div></div>`;
        currentDay = slot.day;
        html += `
        <div class="card border-0 bg-light">
            <div class="card-header bg-primary text-white fw-bold py-2"><i class="bi bi-calendar-day"></i> ${currentDay}</div>
            <div class="list-group list-group-flush border-0">`;
    }
    html += `
        <div class="list-group-item bg-transparent d-flex justify-content-between align-items-center">
            <div>
                <span class="badge bg-secondary me-2"><i class="bi bi-clock"></i> ${slot.start} - ${slot.end}</span>
                <span class="fw-bold text-dark">${slot.subject}</span>
            </div>
        </div>`;
});
if (currentDay !== '') html += `</div></div>`;
console.log(html);
