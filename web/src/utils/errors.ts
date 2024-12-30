import axios from 'axios';

/**
 * Extracts an error message from an Axios error or a generic error.
 * @param error - The error object to process.
 * @returns A string containing the error message.
 */
export function extractErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error) && error.response) {
    return (
      error.response.data || 'Failed to process request. Please try again.'
    );
  }

  return (error as Error)?.message || 'An unexpected error occurred.';
}
