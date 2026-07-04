import { httpClient } from './httpClient'
import type { Booking, BookingListResponse, CreateBookingRequest, SubmitRatingRequest } from './types'

export const bookingsApi = {
  createBooking: (req: CreateBookingRequest) => httpClient.post<Booking>('/bookings', req),
  listBookings: () => httpClient.get<BookingListResponse>('/bookings'),
  getBooking: (bookingId: string) => httpClient.get<Booking>(`/bookings/${bookingId}`),
  cancelBooking: (bookingId: string) => httpClient.post<Booking>(`/bookings/${bookingId}/cancel`),
  submitRating: (bookingId: string, req: SubmitRatingRequest) =>
    httpClient.post<Booking>(`/bookings/${bookingId}/rating`, req),
}
