<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

// --- 1. CONFIG & TYPES ---

const API_BASE = 'http://localhost:8080/api'

interface User {
  token: string
  role: string
}

interface Room {
  id: number
  name: string
  capacity: number
  description: string
}

interface Booking {
  id: number
  roomId: number
  userId: number
  title: string
  startTime: string
  endTime: string
}

// --- 2. STATE ---

// Auth
const token = ref<string | null>(null)
const userRole = ref<string>('')
const loginEmail = ref('')
const loginPassword = ref('')
const loginError = ref('')

// Data
const rooms = ref<Room[]>([])
const bookings = ref<Booking[]>([])
const selectedRoomId = ref<number | null>(null)

// Modal
const showBookingModal = ref(false)
const bookingTitle = ref('')
const bookingStartTime = ref('')
const bookingEndTime = ref('')
const bookingError = ref('')

// --- 3. COMPUTED ---

const selectedRoom = computed(() =>
  rooms.value.find(r => r.id === selectedRoomId.value)
)

const roomBookings = computed(() =>
  bookings.value.filter(b => b.roomId === selectedRoomId.value)
)

// --- 4. FUNCTIONS ---

// --- AUTH ---
const login = async () => {
  try {
    loginError.value = ''
    const res = await fetch(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: loginEmail.value,
        password: loginPassword.value
      })
    })

    if (!res.ok) throw new Error('Login failed')

    const data: User = await res.json()
    token.value = data.token
    userRole.value = data.role

    // Сохраняем в localStorage
    localStorage.setItem('token', data.token)
    localStorage.setItem('role', data.role)

    await loadData()
  } catch (err) {
    loginError.value = 'Invalid credentials'
  }
}

const logout = () => {
  token.value = null
  userRole.value = ''
  localStorage.removeItem('token')
  localStorage.removeItem('role')
  rooms.value = []
  bookings.value = []
  selectedRoomId.value = null
}

// --- DATA FETCHING ---
const loadData = async () => {
  if (!token.value) return

  try {
    const [roomsRes, bookingsRes] = await Promise.all([
      fetch(`${API_BASE}/rooms`, {
        headers: { Authorization: `Bearer ${token.value}` }
      }),
      fetch(`${API_BASE}/bookings`, {
        headers: { Authorization: `Bearer ${token.value}` }
      })
    ])

    if (roomsRes.ok) rooms.value = await roomsRes.json()
    if (bookingsRes.ok) bookings.value = await bookingsRes.json()

    // Выбираем первую комнату по умолчанию
    if (rooms.value.length > 0 && !selectedRoomId.value) {
      selectedRoomId.value = rooms.value[0].id
    }
  } catch (err) {
    console.error('Failed to load data', err)
  }
}

// --- BOOKING ACTIONS ---
const openBookingModal = (hour: number) => {
  const start = new Date()
  start.setHours(hour, 0, 0, 0)
  const end = new Date(start)
  end.setHours(hour + 1, 0, 0, 0)

  // Fix timezone offset for input[type="datetime-local"]
  const toLocalISO = (date: Date) => {
    const offset = date.getTimezoneOffset() * 60000
    return new Date(date.getTime() - offset).toISOString().slice(0, 16)
  }

  bookingStartTime.value = toLocalISO(start)
  bookingEndTime.value = toLocalISO(end)
  bookingTitle.value = ''
  bookingError.value = ''
  showBookingModal.value = true
}

const createBooking = async () => {
  if (!token.value || !selectedRoomId.value) return

  try {
    bookingError.value = ''
    const res = await fetch(`${API_BASE}/bookings`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token.value}`
      },
      body: JSON.stringify({
        roomId: selectedRoomId.value,
        title: bookingTitle.value,
        startTime: new Date(bookingStartTime.value).toISOString(),
        endTime: new Date(bookingEndTime.value).toISOString()
      })
    })

    if (res.status === 409) {
      alert('Already booked! Please choose another time.')
      return
    }

    if (!res.ok) throw new Error('Booking failed')

    await loadData()
    showBookingModal.value = false
  } catch (err) {
    bookingError.value = 'Failed to create booking'
  }
}

const deleteBooking = async (id: number) => {
  if (!confirm('Are you sure you want to delete this reservation?')) return

  try {
    const res = await fetch(`${API_BASE}/bookings/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token.value}` }
    })

    if (!res.ok) throw new Error('Delete failed')

    // 🔥 ФИШКА ТУТ: Мы сразу удаляем бронь из списка на экране
    bookings.value = bookings.value.filter(b => b.id !== id)

  } catch (err) {
    alert('Failed to delete booking')
  }
}

// --- TIMELINE HELPERS ---
const hours = Array.from({ length: 13 }, (_, i) => i + 8) // 08:00 - 20:00

const getBookingPosition = (booking: Booking) => {
  const start = new Date(booking.startTime)
  const end = new Date(booking.endTime)

  const startHour = start.getHours() + start.getMinutes() / 60
  const endHour = end.getHours() + end.getMinutes() / 60

  // 12 hours total (08:00 to 20:00)
  const top = ((startHour - 8) / 12) * 100
  const height = ((endHour - startHour) / 12) * 100

  return {
    top: `${top}%`,
    height: `${height}%`
  }
}

const formatTime = (isoString: string) => {
  return new Date(isoString).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

// --- INIT ---
onMounted(() => {
  const savedToken = localStorage.getItem('token')
  const savedRole = localStorage.getItem('role')
  if (savedToken) {
    token.value = savedToken
    userRole.value = savedRole || ''
    loadData()
  }
})
</script>

<template>
  <div v-if="!token" class="min-h-screen w-full bg-gray-100 flex items-center justify-center p-4">
    <div class="bg-white rounded-xl shadow-xl p-8 w-full max-w-md border border-gray-200">
      <h1 class="text-3xl font-bold text-gray-900 mb-2 text-center">Booking System</h1>
      <p class="text-gray-500 text-center mb-8">Sign in to manage reservations</p>

      <form @submit.prevent="login" class="space-y-6">
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">Email Address</label>
          <input
              v-model="loginEmail"
              type="email"
              placeholder="test@user.com"
              required
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition"
          />
        </div>

        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-2">Password</label>
          <input
              v-model="loginPassword"
              type="password"
              placeholder="••••••••"
              required
              class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition"
          />
        </div>

        <div v-if="loginError" class="p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-200">
          {{ loginError }}
        </div>

        <button
            type="submit"
            class="w-full bg-blue-600 text-white py-3 rounded-lg font-semibold hover:bg-blue-700 transition shadow-md"
        >
          Sign In
        </button>
      </form>
    </div>
  </div>

  <div v-else class="min-h-screen w-full bg-gray-50 flex flex-col">
    <header class="bg-white border-b border-gray-200 sticky top-0 z-20">
      <div class="w-full px-6 h-16 flex justify-between items-center">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold">
            M
          </div>
          <h1 class="text-xl font-bold text-gray-900">Moscow Polytech Booking</h1>

          <span v-if="userRole === 'admin'" class="ml-2 px-2 py-0.5 bg-red-100 text-red-700 text-xs font-bold rounded uppercase">
            Admin
          </span>
        </div>

        <button
            @click="logout"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition"
        >
          Logout
        </button>
      </div>
    </header>

    <main class="flex-1 w-full px-6 py-6 flex flex-col lg:flex-row gap-6 overflow-hidden h-[calc(100vh-64px)]">

      <aside class="w-full lg:w-80 flex-shrink-0 flex flex-col h-full overflow-hidden">
        <h2 class="text-lg font-semibold text-gray-900 mb-4 flex-shrink-0">Available Rooms</h2>
        <div class="space-y-3 overflow-y-auto pr-2 custom-scrollbar flex-1">
          <div
              v-for="room in rooms"
              :key="room.id"
              @click="selectedRoomId = room.id"
              class="group p-4 rounded-xl border transition-all cursor-pointer relative"
              :class="[
              selectedRoomId === room.id
                ? 'bg-blue-50 border-blue-500 shadow-md ring-1 ring-blue-500'
                : 'bg-white border-gray-200 hover:border-blue-300 hover:shadow-sm'
            ]"
          >
            <div class="flex justify-between items-start mb-1">
              <h3 class="font-bold text-gray-900 group-hover:text-blue-700 transition">{{ room.name }}</h3>
              <span class="text-xs font-medium px-2 py-1 bg-gray-100 rounded text-gray-600">
                {{ room.capacity }} ppl
              </span>
            </div>
            <p class="text-sm text-gray-500 line-clamp-2">{{ room.description }}</p>
          </div>
        </div>
      </aside>

      <section class="flex-1 bg-white rounded-2xl shadow-sm border border-gray-200 flex flex-col h-full overflow-hidden">
        <div class="p-6 border-b border-gray-100 flex justify-between items-center bg-white flex-shrink-0">
          <div>
            <h2 class="text-xl font-bold text-gray-900">
              {{ selectedRoom ? selectedRoom.name : 'Select a room' }}
            </h2>
            <p class="text-sm text-gray-500">
              {{ selectedRoom ? 'Click on a time slot to book' : 'Please select a room from the list' }}
            </p>
          </div>
        </div>

        <div v-if="selectedRoom" class="flex-1 relative overflow-y-auto custom-scrollbar p-4">
          <div class="relative h-[800px] border-l border-r border-gray-100 bg-white ml-14 mr-4">

            <div
                v-for="(hour, index) in hours"
                :key="hour"
                class="absolute w-full border-t border-gray-100 group cursor-pointer hover:bg-blue-50/50 transition-colors"
                :style="{ top: `${(index / 12) * 100}%`, height: `${100 / 12}%` }"
                @click="openBookingModal(hour)"
            >
              <span class="absolute -left-14 -top-3 text-xs font-medium text-gray-400 w-10 text-right group-hover:text-blue-600">
                {{ hour }}:00
              </span>
              <div class="hidden group-hover:flex absolute inset-0 items-center justify-center opacity-0 group-hover:opacity-100 pointer-events-none">
                <span class="text-blue-500 text-sm font-semibold">+ Book Slot</span>
              </div>
            </div>

            <div
                v-for="booking in roomBookings"
                :key="booking.id"
                class="absolute left-2 right-2 rounded-lg px-3 py-2 border-l-4 shadow-sm overflow-hidden hover:shadow-md transition-shadow z-10 cursor-default group"
                :class="[
                  'bg-blue-100 border-blue-500 text-blue-900'
                ]"
                :style="getBookingPosition(booking)"
            >
              <button
                v-if="userRole === 'admin'"
                @click.stop="deleteBooking(booking.id)"
                class="absolute top-1 right-1 bg-white/70 hover:bg-red-500 hover:text-white text-red-600 rounded-md w-6 h-6 flex items-center justify-center text-xs font-bold transition opacity-0 group-hover:opacity-100 z-50 cursor-pointer"
                title="Delete Reservation"
              >
                ✕
              </button>

              <div class="font-bold text-sm truncate leading-tight pr-6">{{ booking.title }}</div>
              <div class="text-xs opacity-75 mt-0.5">
                {{ formatTime(booking.startTime) }} - {{ formatTime(booking.endTime) }}
              </div>
            </div>

          </div>
        </div>

        <div v-else class="flex-1 flex items-center justify-center text-gray-400 flex-col gap-4">
          <div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center text-2xl">👈</div>
          <p>Select a room to view schedule</p>
        </div>
      </section>
    </main>

    <div v-if="showBookingModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-gray-900/60 backdrop-blur-sm" @click="showBookingModal = false"></div>

      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-lg relative z-10 overflow-hidden transform transition-all">
        <div class="px-6 py-4 border-b border-gray-100 bg-gray-50 flex justify-between items-center">
          <h3 class="text-lg font-bold text-gray-900">New Reservation</h3>
          <button @click="showBookingModal = false" class="text-gray-400 hover:text-gray-600">✕</button>
        </div>

        <div class="p-6">
          <form @submit.prevent="createBooking" class="space-y-5">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Meeting Title</label>
              <input
                  v-model="bookingTitle"
                  type="text"
                  placeholder="e.g. Weekly Sync"
                  required
                  class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Starts</label>
                <input
                    v-model="bookingStartTime"
                    type="datetime-local"
                    required
                    class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-blue-500 outline-none"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Ends</label>
                <input
                    v-model="bookingEndTime"
                    type="datetime-local"
                    required
                    class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-blue-500 outline-none"
                />
              </div>
            </div>

            <div v-if="bookingError" class="text-red-500 text-sm bg-red-50 p-2 rounded">
              {{ bookingError }}
            </div>

            <div class="flex gap-3 mt-6">
              <button
                  type="button"
                  @click="showBookingModal = false"
                  class="flex-1 px-4 py-2.5 bg-white text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 font-medium transition"
              >
                Cancel
              </button>
              <button
                  type="submit"
                  class="flex-1 px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium transition shadow-sm"
              >
                Confirm Booking
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
/* Custom Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>