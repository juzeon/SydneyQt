import Swal, {SweetAlertResult} from 'sweetalert2'

export interface ISwal {
    success: (text: string) => Promise<SweetAlertResult>,
    error: (text: string) => Promise<SweetAlertResult>,
    confirm: (text: string) => Promise<SweetAlertResult>,
}

export let swal: ISwal = {
    success(text: string) {
        return Swal.fire({
            title: 'Success',
            text,
            icon: 'success'
        })
    },
    error(text: string) {
        return Swal.fire({
            title: 'Error',
            text,
            icon: 'error'
        })
    },
    confirm(text: string) {
        return Swal.fire({
            title: 'Confirmation',
            icon: 'question',
            text: text,
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