import { defineComponent, ref } from 'vue'

// Composable stubs
export function useAuth() {
  return {
    userId: ref<string | null>(null),
    isSignedIn: ref(false),
    getToken: ref<(() => Promise<string | null>) | null>(null),
  }
}

export function useUser() {
  return {
    user: ref(null),
  }
}

export function useClerk() {
  return {
    signOut: async () => {},
  }
}

export const useSignIn = () => ({
  signIn: ref(null),
  setActive: async () => {},
})

export const useSignUp = () => ({
  signUp: ref(null),
  setActive: async () => {},
})

// Component stubs
export const SignIn = defineComponent({
  name: 'SignInStub',
  setup() {
    return () => null
  },
})

export const SignUp = defineComponent({
  name: 'SignUpStub',
  setup() {
    return () => null
  },
})

export const UserButton = defineComponent({
  name: 'UserButtonStub',
  setup() {
    return () => null
  },
})

export const ClerkProvider = defineComponent({
  name: 'ClerkProviderStub',
  setup(_props: any, { slots }: any) {
    return () => slots.default?.()
  },
})
