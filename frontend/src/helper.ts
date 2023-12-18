import Swal, {SweetAlertResult} from 'sweetalert2'
import {generate} from "random-words"

export interface ISwal {
    success: (text: string) => Promise<SweetAlertResult>,
    error: (text: string) => Promise<SweetAlertResult>,
    confirm: (text: string) => Promise<SweetAlertResult>,
}

export let swal: ISwal = {
    success(text: string) {
        return Swal.fire({
            title: 'Success',
            html: plainTextToHTML(text),
            icon: 'success'
        })
    },
    error(text: string) {
        return Swal.fire({
            title: 'Error',
            html: plainTextToHTML(text),
            icon: 'error'
        })
    },
    confirm(text: string) {
        return Swal.fire({
            title: 'Confirmation',
            icon: 'question',
            html: plainTextToHTML(text),
            confirmButtonText: 'Confirm',
            cancelButtonText: 'Cancel',
            showCancelButton: true
        })
    }
}

export interface ChatMessage {
    role: string
    type: string
    message: string
}

export function toChatMessages(context: string): ChatMessage[] {
    context += '\n\n[system](#sydney__placeholder)'
    let matches =
        context.matchAll(/\[(system|user|assistant)]\(#(.*?)\)([\s\S]*?)(?=\n.*?(^\[(system|user|assistant)]\(#.*?\)))/gm)
    return Array.from(matches)
        .filter(v => v[2] !== 'sydney__placeholder').map(v => <ChatMessage>{
            role: v[1],
            type: v[2],
            message: v[3].trim(),
        })
}

export function shadeColor(color: string, percent: number): string {
    let R = parseInt(color.substring(1, 3), 16)
    let G = parseInt(color.substring(3, 5), 16)
    let B = parseInt(color.substring(5, 7), 16)
    R = parseInt(String(R * (100 + percent) / 100))
    G = parseInt(String(G * (100 + percent) / 100))
    B = parseInt(String(B * (100 + percent) / 100))
    R = (R < 255) ? R : 255
    G = (G < 255) ? G : 255
    B = (B < 255) ? B : 255
    R = Math.round(R)
    G = Math.round(G)
    B = Math.round(B)
    let RR = ((R.toString(16).length == 1) ? "0" + R.toString(16) : R.toString(16))
    let GG = ((G.toString(16).length == 1) ? "0" + G.toString(16) : G.toString(16))
    let BB = ((B.toString(16).length == 1) ? "0" + B.toString(16) : B.toString(16))
    return "#" + RR + GG + BB
}

export function generateRandomName() {
    return generate({maxLength: 6}) + '_' + generate({maxLength: 6})
}

export function plainTextToHTML(text: string) {
    return '<p>' + escapeHtml(text.trim()).split('\n').join('</p><p>') + '</p>'
}

export function escapeHtml(unsafe: string) {
    return unsafe
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#039;')
}